package main

import (
	"bytes"
	"sync"
	"text/template"

	"github.com/golang-module/carbon/v2"
	"github.com/sbani/go-humanizer/numbers"
)

type TemplateService struct {
	sync.Mutex
	Services      *map[ServiceID]Service
	Cache         *template.Template
	MasterService *MasterService
	Config        *ConfigurationService

	Logger
	Service
}

type Substitutions struct {
	Time    string
	UserNum string
	NL      string
}

func (t *TemplateService) Init(args map[InitArg]interface{}) (err error) {
	t.Logger = Logger{
		Name: "template",
		ID:   TemplateServiceID,
	}

	var ok bool
	t.Config, ok = args[InitArgConfig].(*ConfigurationService)
	if !ok {
		t.LogAlert("config %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	t.Services, ok = args[InitArgServices].(*map[ServiceID]Service)
	if !ok {
		t.LogAlert("services %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	t.MasterService, ok = (*t.Services)[MasterServiceID].(*MasterService)
	if !ok {
		return ErrorInvalidArgument
	}

	t.Rehash()

	return
}

func (t *TemplateService) Run() {
	// noop
}

func (t *TemplateService) Rehash() {
	if t.Config.Values.Service.Templates.MOTD != "" {
		motd, err := template.New("motd").Parse(t.Config.Values.Service.Templates.MOTD)
		if err != nil {
			t.LogAlert("unable to parse motd template")
			return
		}
		t.Cache = motd
	}
}

func (t *TemplateService) Shutdown() {
	// noop
}

func (t *TemplateService) Get() string {
	if t.Cache != nil {
		out := bytes.NewBuffer([]byte{})

		err := t.Cache.Execute(out, Substitutions{
			Time:    carbon.Now().Format(t.Config.Values.Service.Templates.TimeFormat),
			UserNum: numbers.Ordinalize(len(t.MasterService.DailyStats.UniqueUsers)),
			NL:      "\\n",
		})
		if err != nil {
			return ""
		}

		return out.String()
	}

	return ""
}
