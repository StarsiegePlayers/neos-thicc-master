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
	Config        *Configuration
	Services      *map[ServiceID]Service
	Cache         *template.Template
	MasterService *MasterService

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
	t.Config, ok = args[InitArgConfig].(*Configuration)
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
	motd, err := template.New("motd").Parse(t.Config.Service.Templates.MOTD)
	if err != nil {
		t.LogAlert("unable to parse motd template")
		return
	}
	t.Cache = motd
}

func (t *TemplateService) Shutdown() {
	// noop
}

func (t *TemplateService) Get() string {
	if t.Cache == nil {
		t.Rehash()
	}
	if t.Cache == nil {
		t.LogAlert("template has not been successfully compiled")
		return ""
	}

	return t.perform(t.Cache)
}

func (t *TemplateService) perform(input *template.Template) string {
	out := bytes.NewBuffer([]byte{})

	t.Log("Current number of users: %d", len(t.MasterService.DailyStats.UniqueUsers))

	err := input.Execute(out, Substitutions{
		Time:    carbon.Now().Format(t.Config.Service.Templates.TimeFormat),
		UserNum: numbers.Ordinalize(len(t.MasterService.DailyStats.UniqueUsers)),
		NL:      "\\n",
	})
	if err != nil {
		return ""
	}

	return out.String()
}
