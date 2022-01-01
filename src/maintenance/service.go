package maintenance

import (
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
)

type Service struct {
	*time.Ticker
	time.Duration
	Services *map[service.ID]service.Interface
	Config   *config.Service

	*log.Log

	service.Interface
}

func (s *Service) Init(services *map[service.ID]service.Interface) (err error) {
	s.Services = services
	s.Config = (*s.Services)[service.Config].(*config.Service)
	s.Log = (*s.Services)[service.Logger].(*log.Service).NewLogger(service.Maintenance)

	s.Ticker = time.NewTicker(s.Config.Values.Advanced.Maintenance.Interval.Duration)

	return
}

func (s *Service) Rehash() {
	s.Shutdown()
	s.Run()
}

func (s *Service) Run() {
	s.Logf("will run every %s", s.Config.Values.Advanced.Maintenance.Interval.String())

	for range s.C {
		for _, v := range *s.Services {
			if sv, ok := v.(service.Maintainable); ok {
				sv.Maintenance()
			}
		}
	}
}

func (s *Service) Shutdown() {
	s.Stop()
	s.Logf("shutdown complete")
}

func (s *Service) Get() string {
	return ""
}
