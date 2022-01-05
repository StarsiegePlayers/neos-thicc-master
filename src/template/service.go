package template

import (
	"bytes"
	"sync"
	"text/template"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/master"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"

	"github.com/golang-module/carbon/v2"
	"github.com/sbani/go-humanizer/numbers"
)

type Service struct {
	sync.Mutex
	Services      *map[service.ID]service.Interface
	Cache         *template.Template
	MasterService *master.Service
	Config        *config.Service

	*log.Log
	service.Interface
	service.Acquirable
}

type Substitutions struct {
	Time    string
	UserNum string
	NL      string
}

func (t *Service) Init(services *map[service.ID]service.Interface) (err error) {
	t.Services = services
	t.Config = (*t.Services)[service.Config].(*config.Service)
	t.MasterService = (*t.Services)[service.Master].(*master.Service)
	t.Log = (*t.Services)[service.Log].(*log.Service).NewLogger(service.Template)
	t.Rehash()

	return
}

func (t *Service) Run() {
	// noop
}

func (t *Service) Rehash() {
	t.Logf("reloading templates")
	motd, err := template.New("motd").Parse(t.Config.Values.Service.Templates.MOTD)

	if err != nil {
		t.LogAlertf("unable to parse motd template")
		return
	}

	t.Cache = motd
}

func (t *Service) Shutdown() {
	// noop
}

func (t *Service) Get() string {
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
