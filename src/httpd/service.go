package httpd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/master"
	"github.com/StarsiegePlayers/neos-thicc-master/src/polling"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
)

type Service struct {
	sync.Mutex
	srv    *http.Server
	router *Router

	Services      *map[service.ID]service.Interface
	Config        *config.Service
	MasterService *master.Service
	PollService   *polling.PollService
	STUNService   service.Interface

	listenIP   string
	listenPort uint16

	cache HTTPCache

	Logs struct {
		HTTPD  *log.Log
		Router *log.Log
	}
	service.Interface
	service.Maintainable
}

const (
	ShutdownTimer = 5 * time.Second
)

func (s *Service) Init(services *map[service.ID]service.Interface) (err error) {
	s.Services = services
	s.Config = (*s.Services)[service.Config].(*config.Service)
	s.MasterService = (*s.Services)[service.Master].(*master.Service)
	s.PollService, _ = (*s.Services)[service.Poll].(*polling.PollService)
	s.STUNService = (*s.Services)[service.STUN]
	s.Logs.HTTPD = (*s.Services)[service.Log].(*log.Service).NewLogger(service.HTTPDRouter)
	s.Logs.Router = (*s.Services)[service.Log].(*log.Service).NewLogger(service.HTTPDRouter)

	if s.router == nil {
		s.router = NewHTTPRouter(s.Logs.Router, s.Config.BuildInfo)
	}

	s.registerRoutes()

	s.cache = make(map[HTTPCacheID]interface{})
	s.cache[cacheAdminSessions] = make(map[string]*HTTPAdminSession)
	s.cache[cacheThrottle] = make(map[string]int)
	s.cache[cacheMultiplayer] = make(map[string]*CacheResponse)

	s.srv = s.newServer()

	return nil
}

func (s *Service) Maintenance() {
	go s.maintenanceMultiplayerServersCache()
	s.clearThrottleCache()
}

func (s *Service) newServer() (out *http.Server) {
	s.listenIP = s.Config.Values.HTTPD.Listen.IP
	if s.listenIP == "" {
		s.listenIP = s.Config.Values.Service.Listen.IP
	}

	s.listenPort = s.Config.Values.HTTPD.Listen.Port
	if s.listenPort <= 0 {
		s.listenPort = s.Config.Values.Service.Listen.Port
	}

	ipPort := fmt.Sprintf("%s:%d", s.listenIP, s.listenPort)
	out = &http.Server{
		Addr:    ipPort,
		Handler: s.router.Mux(),
	}

	return
}

func (s *Service) Run() {
	ip := s.listenIP
	if ip == "" {
		ip = service.LocalhostAddress
	}

	localIPPort := fmt.Sprintf("%s:%d", ip, s.listenPort)
	externalIPPort := fmt.Sprintf("%s:%d", s.STUNService.Get(), s.listenPort)

	s.Logs.HTTPD.Logf("now listening on http://%s/ | http://%s/", externalIPPort, localIPPort)

	if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		s.Logs.HTTPD.LogAlertf("error during listen [%w]", err)
	}
}

func (s *Service) Rehash() {
	listenAddr := fmt.Sprintf("%s:%d", s.Config.Values.Service.Listen.IP, s.Config.Values.Service.Listen.Port)
	if s.srv.Addr != listenAddr {
		s.Shutdown()
	}
}

func (s *Service) Shutdown() {
	ctxShutDown, cancel := context.WithTimeout(context.Background(), ShutdownTimer)
	defer func() {
		cancel()
	}()

	if err := s.srv.Shutdown(ctxShutDown); err != nil {
		s.Logs.HTTPD.LogAlertf("shutdown failed: %s", err)
		return
	}

	s.Logs.HTTPD.Logf("shutdown complete")
}
