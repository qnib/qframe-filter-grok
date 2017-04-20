package qframe_filter_grok

import (
	"C"
	"log"
	"fmt"
	"reflect"
	"github.com/vjeantet/grok"
	"github.com/zpatrick/go-config"

	"github.com/qnib/qframe-types"
	"github.com/qnib/qframe-utils"
)

const (
	version = "0.1.3"
	pluginTyp = "filter"
)

type Plugin struct {
	qtypes.Plugin
	grok *grok.Grok
	pattern string
}

func New(qChan qtypes.QChan, cfg config.Config, name string) (p Plugin, err error) {
	p = Plugin{
		Plugin: qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, name, version),
	}
	p.grok, _ = grok.New()
	pCfg := fmt.Sprintf("filter.%s.pattern", p.Name)
	p.pattern, err = p.Cfg.String(pCfg)
	if err != nil {
		log.Printf("[EE] Could not find pattern in config: '%s'", pCfg)
		return p, err
	}
	return p, err
}


func (p *Plugin) Match(str string) (map[string]string, bool) {
	match := true
	val, _ := p.grok.Parse(p.pattern, str)
	keys := reflect.ValueOf(val).MapKeys()
	if len(keys) == 0 {
		match = false
	}
	return val, match
}

func (p *Plugin) GetPattern() (string) {
	return p.pattern
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("info", fmt.Sprintf("Start grok filter v%s", p.Version))
	myId := qutils.GetGID()
	bg := p.QChan.Data.Join()
	inputs := p.GetInputs()
	srcSuccess := p.CfgBoolOr("source-success", true)
	for {
		val := bg.Recv()
		switch val.(type) {
		case qtypes.QMsg:
			qm := val.(qtypes.QMsg)
			if qm.SourceID == myId {
				continue
			}
			if len(inputs) != 0 && !qutils.IsInput(inputs, qm.Source) {
				continue
			}
			if qm.SourceSuccess != srcSuccess {
				continue
			}
			qm.Type = "filter"
			qm.Source = p.Name
			qm.SourceID = myId
			qm.SourcePath = append(qm.SourcePath, p.Name)
			qm.KV, qm.SourceSuccess = p.Match(qm.Msg)
			p.QChan.Data.Send(qm)
		}
	}
}
