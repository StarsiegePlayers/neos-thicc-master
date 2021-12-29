package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type HTTPDService struct {
	sync.Mutex
	srv    *http.Server
	router *Router

	Services      *map[ServiceID]Service
	Config        *ConfigurationService
	MasterService *MasterService
	PollService   *PollService

	listenIP   string
	listenPort uint16

	cache map[HTTPCacheID]interface{}

	Service
	Maintainable
	Logger
}

type HTTPCacheID int

const (
	HTTPCacheMultiplayer = HTTPCacheID(iota)
	HTTPCacheAdminSessions
	HTTPCacheThrottle
)

var (
	//go:embed www-build/*
	wwwFS embed.FS
)

func (s *HTTPDService) Init(args map[InitArg]interface{}) (err error) {
	s.Logger = Logger{
		Name: "HTTP Server",
		ID:   HTTPServiceID,
	}

	var ok bool

	s.Config, ok = args[InitArgConfig].(*ConfigurationService)
	if !ok {
		s.LogAlert("config %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	s.Services, ok = args[InitArgServices].(*map[ServiceID]Service)
	if !ok {
		s.LogAlert("service %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	s.MasterService, ok = (*s.Services)[MasterServiceID].(*MasterService)
	if !ok {
		return ErrorInvalidArgument
	}

	s.PollService, ok = (*s.Services)[PollServiceID].(*PollService)
	if !ok {
		// gracefully handle a disabled poll service
		s.PollService = nil
	}

	if s.router == nil {
		s.router = NewHttpRouter()
	}
	s.registerRoutes()

	s.cache = make(map[HTTPCacheID]interface{})
	s.cache[HTTPCacheAdminSessions] = make(map[string]*HTTPAdminSession, 0)
	s.cache[HTTPCacheThrottle] = make(map[string]int)

	s.srv = s.newServer()
	return nil
}

func (s *HTTPDService) Maintenance() {
	go s.maintenanceMultiplayerServersCache()
	s.clearThrottleCache()
}

func (s *HTTPDService) newServer() (out *http.Server) {
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

func (s *HTTPDService) Run() {
	ip := s.listenIP
	if ip == "" {
		ip = "127.0.0.1"
	}

	localIpPort := fmt.Sprintf("%s:%d", ip, s.listenPort)
	externalIpPort := fmt.Sprintf("%s:%d", s.Config.Values.externalIP, s.listenPort)

	s.Log("now listening on http://%s/ | http://%s/", externalIpPort, localIpPort)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.LogAlert("error during listen %s", err)
	}
}

func (s *HTTPDService) Rehash() {
	listenAddr := fmt.Sprintf("%s:%d", s.Config.Values.Service.Listen.IP, s.Config.Values.Service.Listen.Port)
	if s.srv.Addr != listenAddr {
		s.Shutdown()
	}
}

func (s *HTTPDService) Shutdown() {
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := s.srv.Shutdown(ctxShutDown); err != nil {
		s.LogAlert("shutdown failed: %s", err)
		return
	}
	s.Log("shutdown complete")
}
