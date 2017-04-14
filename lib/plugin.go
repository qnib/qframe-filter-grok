package qframe_filter_grok

import (
	"C"
	"log"
	"fmt"
	"reflect"
	"strings"

	"github.com/vjeantet/grok"
	"github.com/zpatrick/go-config"

	"github.com/qnib/qframe-types"
	"github.com/qnib/qframe-utils"
)

const (
	version = "0.1.0"
)

type Plugin struct {
	qtypes.Plugin
	grok *grok.Grok
	pattern string
}

func New(qChan qtypes.QChan, cfg config.Config, name string) (p Plugin, err error) {
	p = Plugin{
		Plugin: qtypes.Plugin{
			QChan: qChan,
			Cfg:   cfg,
		},
	}
	p.Version = version
	p.Name = name
	p.grok, _ = grok.New()
	pCfg := fmt.Sprintf("filter.%s.pattern", p.Name)
	p.pattern, err = cfg.String(pCfg)
	if err != nil {
		log.Printf("[EE] Could not find pattern in config: '%s'", pCfg)
		return p, err
	}
	return p, err
}


func (p *Plugin) Match(str string) (val map[string]string, err error) {
	val, err = p.grok.Parse(p.pattern, str)
	keys := reflect.ValueOf(val).MapKeys()
	if len(keys) == 0 {
		fmt.Println("Sorry, not a single match...")
	}
	return val, err
}

func (p *Plugin) GetPattern() (string) {
	return p.pattern
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	log.Printf("[II] Start grok filter '%s' v%s", p.Name, p.Version)
	myId := qutils.GetGID()
	bg := p.QChan.Data.Join()
	for {
		val := bg.Recv()
		switch val.(type) {
		case qtypes.QMsg:
			qm := val.(qtypes.QMsg)
			if qm.SourceID == myId {
				continue
			}
			qm.Type = "filter"
			qm.Source = strings.Join(append(strings.Split(qm.Source, "->"), "id"), "->")
			qm.SourceID = myId
			// GROK magic
			//var err error
			qm.KV, _ = p.Match(qm.Msg)
			/*if err != nil {
				continue
			}*/
			p.QChan.Data.Send(qm)
		}
	}
}
