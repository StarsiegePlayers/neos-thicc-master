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

	services struct {
		Map    *map[service.ID]service.Interface
		Config *config.Service
	}

	logs struct {
		Maintenance *log.Log
	}

	status service.LifeCycle

	service.Interface
	service.Runnable
}

func (s *Service) Init(services *map[service.ID]service.Interface) (err error) {
	s.services.Map = services
	s.services.Config = (*s.services.Map)[service.Config].(*config.Service)
	s.logs.Maintenance = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.Maintenance)

	s.Duration = s.services.Config.Values.Advanced.Maintenance.Interval.Duration
	s.status = service.Starting

	return
}

func (s *Service) Status() service.LifeCycle {
	return s.status
}

func (s *Service) Run() {
	s.logs.Maintenance.Logf("will run every %s", s.services.Config.Values.Advanced.Maintenance.Interval.String())
	s.status = service.Running

	s.Duration = s.services.Config.Values.Advanced.Maintenance.Interval.Duration
	s.Ticker = time.NewTicker(s.Duration)

	for range s.C {
		for _, v := range *s.services.Map {
			if sv, ok := v.(service.Maintainable); ok {
				sv.Maintenance()
			}
		}
	}

	s.status = service.Stopped
	s.logs.Maintenance.LogAlertf("service stopped")
}

func (s *Service) Rehash() {
	p := s.status
	s.status = service.Rehashing

	if s.Duration != s.services.Config.Values.Advanced.Maintenance.Interval.Duration {
		if p == service.Running {
			s.Shutdown()
		}

		go s.Run()
	}

	s.status = p
}

func (s *Service) Shutdown() {
	s.status = service.Stopping
	s.Stop()
	s.status = service.Stopped
	s.logs.Maintenance.Logf("shutdown complete")
}
