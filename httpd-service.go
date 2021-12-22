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

	Config   *Configuration
	Services *map[string]Service

	listenIP   string
	listenPort uint16

	Service
	Logger
}

var (
	//go:embed www-build/*
	wwwFS embed.FS
)

func (s *HTTPDService) Init(args map[string]interface{}) (err error) {
	s.Logger = Logger{
		Name:   "HTTP Server",
		LogTag: "httpd",
	}

	var ok bool
	s.Services, ok = args["services"].(*map[string]Service)
	if !ok {
		s.LogAlert("services %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	s.Config, ok = args["config"].(*Configuration)
	if !ok {
		s.LogAlert("config %s", ErrorInvalidArgument)
		return ErrorInvalidArgument
	}

	if s.router == nil {
		s.router = NewHttpRouter("/api")
	}
	s.registerRoutes()

	s.srv = s.newServer()
	return nil
}

func (s *HTTPDService) newServer() (out *http.Server) {
	s.listenIP = s.Config.HTTPD.Listen.IP
	if s.listenIP == "" {
		s.listenIP = s.Config.Service.Listen.IP
	}

	s.listenPort = s.Config.HTTPD.Listen.Port
	if s.listenPort <= 0 {
		s.listenPort = s.Config.Service.Listen.Port
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
	externalIpPort := fmt.Sprintf("%s:%d", s.Config.externalIP, s.listenPort)

	s.Log("now listening on http://%s/ | http://%s/", externalIpPort, localIpPort)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.LogAlert("error during listen %s", err)
	}
}

func (s *HTTPDService) Rehash() {
	listenAddr := fmt.Sprintf("%s:%d", s.Config.Service.Listen.IP, s.Config.Service.Listen.Port)
	if s.srv.Addr != listenAddr {
		s.Shutdown()
	}
}

func (s *HTTPDService) Shutdown() {
	s.Log("shutdown requested")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := s.srv.Shutdown(ctxShutDown); err != nil {
		s.LogAlert("shutdown failed: %s", err)
	}
}
