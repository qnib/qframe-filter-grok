package main

import (
	"fmt"
	"testing"
	"time"
	"log"
	"github.com/zpatrick/go-config"

	"github.com/qnib/qframe-types"
	"github.com/qnib/qframe-filter-grok/lib"
	"github.com/stretchr/testify/assert"
	"reflect"
)

func Receive(qchan qtypes.QChan, endCnt int) {
	bg := qchan.Data.Join()
	allCnt := 1
	cnt := 1
	for {
		select {
		case val := <-bg.Read:
			allCnt++
			switch val.(type) {
			case qtypes.Message:
				qm := val.(qtypes.Message)
				if qm.IsLastSource("grok") {
					cnt++
				}
			default:
				fmt.Printf("received msg %d: type=%s\n", allCnt, reflect.TypeOf(val))

			}
		}
		if endCnt == cnt {
			qchan.Data.Send(cnt)
			break
		}
	}
}

func BenchmarkGrok(b *testing.B) {
	endCnt := 199999
	qChan := qtypes.NewQChan()
	qChan.Broadcast()
	cfgMap := map[string]string{
		"log.level": "info",
		"filter.grok.pattern": "test%{INT:number}",
		"filter.grok.inputs": "test",
		"filter.grok.pattern-dir": "/usr/local/src/github.com/qnib/qframe-filter-grok/resources/patterns",
	}
	go Receive(qChan, endCnt)
	cfg := config.NewConfig([]config.Provider{config.NewStatic(cfgMap)})
	p, err := qframe_filter_grok.New(qChan, cfg, "grok")
	if err != nil {
		log.Printf("[EE] Failed to create filter: %v", err)
		return
	}
	dc := qChan.Data.Join()
	go p.Run()
	time.Sleep(time.Duration(50)*time.Millisecond)
	p.Log("info", fmt.Sprintf("Benchmark sends %d messages to grok", endCnt))
	for i := 1; i <= endCnt; i++ {
		msg := fmt.Sprintf("test%d", i)
		qm := qtypes.NewMessage(qtypes.NewBase("test"), "test", "testMsg", msg)
		qm.Message = msg
		qChan.Data.Send(qm)
	}
	done := false
	for {
		select {
		case val := <- dc.Read:
			switch val.(type) {
			case int:
				vali := val.(int)
				assert.Equal(b, endCnt, vali)
				done = true
			}
		case <-time.After(5 * time.Second):
				b.Fatal("metrics receive timeout")
		}
		if done {
			break
		}
	}
}


