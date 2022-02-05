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
	"github.com/StarsiegePlayers/neos-thicc-master/src/stun"
)

type Service struct {
	sync.Mutex
	srv    *http.Server
	router *Router

	listenIP   string
	listenPort uint16
	status     service.LifeCycle
	cache      HTTPCache

	services struct {
		Map      *map[service.ID]service.Interface
		Config   *config.Service
		Master   *master.Service
		Poll     *polling.Service
		STUN     *stun.Service
		Template service.Getable
	}
	logs struct {
		HTTPD  *log.Log
		Router *log.Log
	}

	service.Interface
	service.Runnable
	service.Maintainable
}

const (
	ShutdownTimer = 5 * time.Second
)

func (s *Service) Init(services *map[service.ID]service.Interface) (err error) {
	s.status = service.Starting
	s.services.Map = services
	s.services.Config = (*s.services.Map)[service.Config].(*config.Service)
	s.services.Master = (*s.services.Map)[service.Master].(*master.Service)
	s.services.Poll, _ = (*s.services.Map)[service.Poll].(*polling.Service)
	s.services.STUN = (*s.services.Map)[service.STUN].(*stun.Service)
	s.services.Template = (*s.services.Map)[service.Template].(service.Getable)
	s.logs.HTTPD = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.HTTPDRouter)
	s.logs.Router = (*s.services.Map)[service.Log].(*log.Service).NewLogger(service.HTTPDRouter)

	if s.router == nil {
		s.router = NewHTTPRouter(s.logs.Router, s.services.Config.BuildInfo, s.services.Config)
	}

	s.cache = make(map[HTTPCacheID]interface{})
	s.cache[cacheAdminSessions] = make(map[string]*HTTPAdminSession)
	s.cache[cacheThrottle] = make(map[string]int)
	s.cache[cacheMultiplayer] = make(map[string]*CacheResponse)

	s.registerRoutes()

	return
}

func (s *Service) newServer() (out *http.Server) {
	s.listenIP = s.services.Config.Values.HTTPD.Listen.IP
	if s.listenIP == "" {
		s.listenIP = s.services.Config.Values.Service.Listen.IP
	}

	s.listenPort = s.services.Config.Values.HTTPD.Listen.Port
	if s.listenPort <= 0 {
		s.listenPort = s.services.Config.Values.Service.Listen.Port
	}

	ipPort := fmt.Sprintf("%s:%d", s.listenIP, s.listenPort)
	out = &http.Server{
		Addr:    ipPort,
		Handler: s.router.Mux(),
	}

	return
}

func (s *Service) Status() service.LifeCycle {
	return s.status
}

func (s *Service) Maintenance() {
	s.maintenanceMultiplayerServersCache()
	s.clearThrottleCache()
}

func (s *Service) Run() {
	s.srv = s.newServer()

	ip := s.listenIP
	if ip == "" {
		ip = service.LocalhostAddress
	}

	localIPPort := fmt.Sprintf("%s:%d", ip, s.listenPort)
	externalIPPort := fmt.Sprintf("%s:%d", s.services.STUN.Get(""), s.listenPort)

	s.logs.HTTPD.Logf("now listening on http://%s/ | http://%s/", externalIPPort, localIPPort)

	s.status = service.Running
	if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		s.logs.HTTPD.LogAlertf("error during listen [%w]", err)
	}
}

func (s *Service) Rehash() {
	var p service.LifeCycle
	p, s.status = s.status, service.Rehashing

	// invalidate our caches, except for the admin logins
	admins := s.cache[cacheAdminSessions]

	s.Lock()
	s.cache = make(map[HTTPCacheID]interface{})
	s.cache[cacheThrottle] = make(map[string]int)
	s.cache[cacheMultiplayer] = make(map[string]*CacheResponse)
	s.cache[cacheAdminSessions] = admins
	s.Unlock()

	s.status = p
}

func (s *Service) Shutdown() {
	s.status = service.Stopping

	ctxShutDown, cancel := context.WithTimeout(context.Background(), ShutdownTimer)
	defer cancel()

	if err := s.srv.Shutdown(ctxShutDown); err != nil {
		s.logs.HTTPD.LogAlertf("shutdown failed: %s", err)
		return
	}

	s.status = service.Stopped
	s.logs.HTTPD.Logf("shutdown complete")
}
