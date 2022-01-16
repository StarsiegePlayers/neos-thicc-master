package src

import (
	"context"
	"sort"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/httpd"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/maintenance"
	"github.com/StarsiegePlayers/neos-thicc-master/src/master"
	"github.com/StarsiegePlayers/neos-thicc-master/src/polling"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
	"github.com/StarsiegePlayers/neos-thicc-master/src/stats"
	"github.com/StarsiegePlayers/neos-thicc-master/src/stun"
	"github.com/StarsiegePlayers/neos-thicc-master/src/template"
)

type Server struct {
	context.Context
	cancel context.CancelFunc

	Services map[service.ID]service.Interface

	Logs struct {
		startup  *log.Log
		shutdown *log.Log
		restart  *log.Log
		rehash   *log.Log
	}

	status service.LifeCycle

	service.Interface
	service.Runnable
}

func (s *Server) Init(*map[service.ID]service.Interface) error {
	s.Context, s.cancel = context.WithCancel(context.Background())
	s.status = service.Starting

	s.Services = make(map[service.ID]service.Interface)

	loggerService := new(log.Service)
	_ = loggerService.Init(&s.Services)
	s.Services[service.Log] = loggerService

	configService := new(config.Service)
	_ = configService.Init(&s.Services)
	s.Services[service.Config] = configService

	s.Services[service.Template] = new(template.Service)
	s.Services[service.Stats] = new(stats.Service)
	s.Services[service.Master] = new(master.Service)
	s.Services[service.Maintenance] = new(maintenance.Service)
	s.Services[service.DailyMaintenance] = new(maintenance.DailyService)
	s.Services[service.STUN] = new(stun.Service)

	if configService.Values.HTTPD.Enabled {
		s.Services[service.HTTPD] = new(httpd.Service)
	}

	if configService.Values.Poll.Enabled {
		s.Services[service.Poll] = new(polling.Service)
	}

	s.Logs.startup = loggerService.NewLogger(service.Startup)
	s.Logs.rehash = loggerService.NewLogger(service.Rehash)
	s.Logs.shutdown = loggerService.NewLogger(service.Shutdown)
	s.Logs.restart = loggerService.NewLogger(service.Restart)

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

		if sv, ok := s.Services[v].(service.Runnable); ok {
			go sv.Run()
		}
	}

	s.Services[service.Config].(*config.Service).Callback.Rehash = s.Rehash
	s.Services[service.Config].(*config.Service).Callback.StartStopServices = s.startStopServices
	s.Services[service.Config].(*config.Service).Callback.Shutdown = s.Shutdown
	s.Services[service.Config].(*config.Service).Callback.Restart = s.Restart
	s.status = service.Running

	s.Logs.startup.Logf("startup complete")

	// block until shutdown
	<-s.Context.Done()
}

func (s *Server) startStopServices() {
	c := s.Services[service.Config].(*config.Service)

	s.checkStartStopService(c.Values.HTTPD.Enabled, service.HTTPD)
	s.checkStartStopService(c.Values.Poll.Enabled, service.Poll)
}

func (s *Server) Rehash() {
	p := s.status
	s.status = service.Rehashing
	s.Logs.rehash.Logf("rehashing services")

	for _, v := range s.Services {
		if sv, ok := v.(service.Rehashable); ok {
			sv.Rehash()
		}
	}

	s.status = p
}

func (s *Server) Shutdown() {
	s.status = service.Stopping

	for _, v := range s.Services {
		if sv, ok := v.(service.Runnable); ok {
			sv.Shutdown()
		}
	}

	s.status = service.Stopped

	s.cancel()
	s.Logs.shutdown.Logf("shutdown complete")
}

func (s *Server) Restart() {
	s.Logs.restart.Logf("shutting down services")

	for _, v := range s.Services {
		if sv, ok := v.(service.Runnable); ok {
			sv.Shutdown()
		}
	}

	s.Logs.restart.Logf("services down")
	s.Logs.restart.Logf("starting services")

	for _, v := range s.Services {
		if sv, ok := v.(service.Runnable); ok {
			go sv.Run()
		}
	}

	s.Logs.restart.Logf("services started")
}

func (s *Server) Status() service.LifeCycle {
	return s.status
}

// checkStartStopService performs the following:
// 1. checks to see if a particular service is running (and in our services list)
// 2. if it is running, and it shouldn't be, stop it
// 3. if it isn't running, and it should be running, start it
func (s *Server) checkStartStopService(isEnabled bool, id service.ID) {
	s1, found := s.Services[id]
	if isEnabled && !found {
		sv := new(polling.Service)

		err := sv.Init(&s.Services)
		if err != nil {
			s.Logs.rehash.LogAlertf("error starting %s [%w]", id, err)
			return
		}

		go sv.Run()

		s.Logs.rehash.Logf("started %s successfully", id)
		s.Services[id] = sv
	} else if !isEnabled && found {
		if s2, ok2 := s1.(service.Runnable); ok2 {
			s.Logs.rehash.Logf("shutting down %s", id)
			s2.Shutdown()
			delete(s.Services, id)
		}
	}
}
