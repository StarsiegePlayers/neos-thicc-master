package main

import (
	"io/fs"
	"net/http"
)

func (s *HTTPDService) registerRoutes() {
	s.router.SetFileSystem(fs.Sub(wwwFS, "www-build"))
	s.router.AddRoute("/api/v1/admin/serversettings", http.MethodGet, s.middlewareAuth(s.routeGetAdminServerSettings))
	s.router.AddRoute("/api/v1/admin/serversettings", http.MethodPut, s.middlewareAuth(s.routePutAdminServerSettings))
	s.router.AddRoute("/api/v1/admin/login", http.MethodPost, s.middlewareThrottle(s.routePostAdminLogin))
	s.router.AddRoute("/api/v1/multiplayer/servers", http.MethodGet, s.routeGetMultiplayerServers)
}

func (s *HTTPDService) middlewareAuth(fn http.HandlerFunc) http.HandlerFunc {
	return fn
}

func (s *HTTPDService) middlewareThrottle(fn http.HandlerFunc) http.HandlerFunc {
	return fn
}

func (s *HTTPDService) routeGetMultiplayerServers(w http.ResponseWriter, r *http.Request) {

}

func (s *HTTPDService) routePostAdminLogin(w http.ResponseWriter, r *http.Request) {

}

func (s *HTTPDService) routePutAdminServerSettings(w http.ResponseWriter, r *http.Request) {

}

func (s *HTTPDService) routeGetAdminServerSettings(w http.ResponseWriter, r *http.Request) {

}
