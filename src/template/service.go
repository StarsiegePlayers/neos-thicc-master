package template

import (
	"bytes"
	"sync"
	"text/template"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/master"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
	"github.com/StarsiegePlayers/neos-thicc-master/src/stats"

	"github.com/golang-module/carbon/v2"
	"github.com/sbani/go-humanizer/numbers"
)

type Service struct {
	sync.Mutex

	services struct {
		Map    *map[service.ID]service.Interface
		Master *master.Service
		Config *config.Service
		Stats  *stats.Service
	}

	status        service.LifeCycle
	templateCache *template.Template
	log           *log.Log

	service.Interface
	service.Rehashable
	service.Getable
}

type Substitutions struct {
	Time          string
	UserNum       string
	UserTotal     int
	ActiveServers int
	TotalServers  int
	IP            string
	NL            string
}

type SubstitutionParameters struct {
	Host string
	Port string
}

func (s *Service) Init(services *map[service.ID]service.Interface) (err error) {
	s.services.Map = services
	s.services.Config = (*s.services.Map)[service.Config].(*config.Service)
	s.services.Master = (*s.services.Map)[service.Master].(*master.Service)
	s.services.Stats = (*s.services.Map)[service.Stats].(*stats.Service)
	s.log = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.Template)
	s.Rehash()

	s.status = service.Starting

	return
}

func (s *Service) Rehash() {
	p := s.status
	s.status = service.Rehashing

	s.log.Logf("reloading templates")
	motd, err := template.New("motd").Parse(s.services.Config.Values.Service.Templates.MOTD)

	if err != nil {
		s.log.LogAlertf("unable to parse motd template")
		return
	}

	s.templateCache = motd
	s.status = p
}

func (s *Service) Status() service.LifeCycle {
	return s.status
}

func (s *Service) Get(host string) string {
	if s.templateCache != nil {
		out := bytes.NewBuffer([]byte{})

		err := s.templateCache.Execute(out, Substitutions{
			Time:          carbon.Now().Format(s.services.Config.Values.Service.Templates.TimeFormat),
			UserNum:       numbers.Ordinalize(s.services.Stats.GetDailyHostNumber(host)),
			UserTotal:     s.services.Stats.GetDailyClientsTotal(),
			ActiveServers: s.services.Stats.GetTotalServersWithPlayers(),
			TotalServers:  len(s.services.Master.ServerList),
			IP:            host,
			NL:            "\\n",
		})
		if err != nil {
			return ""
		}

		return out.String()
	}

	return ""
}
