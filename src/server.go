package src

import (
	"context"
	"fmt"
	"sort"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/httpd"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/maintenance"
	"github.com/StarsiegePlayers/neos-thicc-master/src/master"
	"github.com/StarsiegePlayers/neos-thicc-master/src/polling"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
	"github.com/StarsiegePlayers/neos-thicc-master/src/stun"
	"github.com/StarsiegePlayers/neos-thicc-master/src/template"
)

type Server struct {
	context.Context
	cancel context.CancelFunc

	Services map[service.ID]service.Interface

	IsRunning bool

	Logs struct {
		startup  *log.Log
		shutdown *log.Log
		rehash   *log.Log
	}

	service.Interface
}

func (s *Server) Init(*map[service.ID]service.Interface) error {
	s.Context, s.cancel = context.WithCancel(context.Background())
	s.IsRunning = true

	s.Services = make(map[service.ID]service.Interface)

	loggerService := new(log.Service)
	_ = loggerService.Init(&s.Services)
	s.Services[service.Log] = loggerService

	configService := new(config.Service)
	_ = configService.Init(&s.Services)
	s.Services[service.Config] = configService

	s.Services[service.Template] = new(template.Service)
	s.Services[service.Master] = new(master.Service)
	s.Services[service.Maintenance] = new(maintenance.Service)
	s.Services[service.DailyMaintenance] = new(maintenance.DailyService)
	s.Services[service.STUN] = new(stun.Service)

	if configService.Values.HTTPD.Enabled {
		s.Services[service.HTTPD] = new(httpd.Service)
	}

	if configService.Values.Poll.Enabled {
		s.Services[service.Poll] = new(polling.PollService)
	}

	s.Logs.startup = loggerService.NewLogger(service.Startup)
	s.Logs.rehash = loggerService.NewLogger(service.Rehash)
	s.Logs.shutdown = loggerService.NewLogger(service.Shutdown)

	s.Logs.startup.Logf("initialization completed")

	return nil
}

func (s *Server) Run() {
	var err error

	serviceList := make(service.IDs, 0)
	for k := range s.Services {
		serviceList = append(serviceList, k)
	}

	sort.Sort(serviceList)

	s.Logs.startup.Logf("starting services")

	for _, v := range serviceList {
		err = s.Services[v].Init(&s.Services)
		if err != nil {
			s.Logs.startup.LogAlertf("service {%s} failed to initialize, removing from threads list - [%s]", v, err)
			delete(s.Services, v)

			continue
		}

		go s.Services[v].Run()
	}

	s.Services[service.Config].(*config.Service).RehashFn = s.Rehash
	s.Services[service.Config].(*config.Service).UpdateRunningServicesFn = s.updateRunningServices

	// block until shutdown
	<-s.Context.Done()
}

func (s *Server) updateRunningServices() {
	var (
		err error
		id  service.ID
	)

	c := s.Services[service.Config].(*config.Service)
	if c.Values.HTTPD.Enabled {
		id = service.HTTPD
		if _, ok := s.Services[id]; !ok {
			sv := new(httpd.Service)
			err = sv.Init(&s.Services)

			if err != nil {
				s.Logs.rehash.LogAlertf("error starting %s [%w]", id, err)
			} else {
				go sv.Run()
				s.Logs.rehash.Logf("started %s successfully", id)
				s.Services[id] = sv
			}
		}
	} else {
		id = service.HTTPD
		if sv, ok := s.Services[id]; ok {
			s.Logs.rehash.Logf("shutting down %s", id)
			sv.Shutdown()
			delete(s.Services, id)
		}
	}
	if c.Values.Poll.Enabled {
		id = service.Poll
		if _, ok := s.Services[id]; !ok {
			sv := new(polling.PollService)
			err = sv.Init(&s.Services)

			if err != nil {
				s.Logs.rehash.LogAlertf("error starting %s [%w]", id, err)
			} else {
				go sv.Run()
				s.Logs.rehash.Logf("started %s successfully", id)
				s.Services[id] = sv
			}
		}
	} else {
		id = service.Poll
		if sv, ok := s.Services[id]; ok {
			s.Logs.rehash.Logf("shutting down %s", id)
			sv.Shutdown()
			delete(s.Services, id)
		}
	}
}

func (s *Server) Rehash() {
	s.Logs.rehash.Logf("reloading services")

	for _, v := range s.Services {
		v.Rehash()
	}
}

func (s *Server) Shutdown() {
	s.IsRunning = false
	for _, v := range s.Services {
		v.Shutdown()
	}

	s.cancel()
	s.Logs.shutdown.Logf("shutdown complete")
}

func (s *Server) Get() string {
	return fmt.Sprintf("%t", s.IsRunning)
}
