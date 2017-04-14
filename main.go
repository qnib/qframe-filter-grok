package main

import (
	"log"
	"fmt"
	"time"

	"github.com/zpatrick/go-config"
	"github.com/qnib/qframe-types"
	"github.com/qnib/qframe-filter-grok/lib"
)

func Run(qChan qtypes.QChan, cfg config.Config, name string) {
	p, _ := qframe_filter_grok.New(qChan, cfg, name)
	p.Run()
}

func main() {
	qChan := qtypes.NewQChan()
	qChan.Broadcast()
	cfgMap := map[string]string{
		"filter.test.pattern": "%{INT:number}",
		"filter.test.inputs": "test",
	}

	cfg := config.NewConfig(
		[]config.Provider{
			config.NewStatic(cfgMap),
		},
	)
	p, err := qframe_filter_grok.New(qChan, *cfg, "test")
	if err != nil {
		log.Printf("[EE] Failed to create filter: %v", err)
		return
	}
	go p.Run()
	time.Sleep(2*time.Second)
	bg := qChan.Data.Join()
	qm := qtypes.NewQMsg("test", "test")
	qm.Msg = "1"
	log.Println("Send message")
	qChan.Data.Send(qm)
	for {
		qm = bg.Recv().(qtypes.QMsg)
		if qm.Source == "test" {
			continue
		}
		fmt.Printf("#### Received result from grok (pattern:%s) filter for input: %s\n", p.GetPattern(), qm.Msg)
		for k, v := range qm.KV {
			fmt.Printf("%+15s: %s\n", k, v)
		}
		break

	}
}
