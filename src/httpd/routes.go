package httpd

import (
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
)

const HTTPStatusEnhanceYourCalm = 420

func (s *Service) registerRoutes() {
	s.router.SetFileSystem(fs.Sub(s.Config.BuildInfo.EmbedFS, "www-build"))
	s.router.AddRoute("/api/v1/multiplayer/servers", http.MethodGet, http.HandlerFunc(s.routeGetMultiplayerServers))
	s.router.AddRoute("/api/v1/admin/login", http.MethodGet, s.middlewareThrottle(s.routeGetAdminLogin))
	s.router.AddRoute("/api/v1/admin/login", http.MethodPost, s.middlewareThrottle(s.routePostAdminLogin))
	s.router.AddRoute("/api/v1/admin/login", http.MethodDelete, s.middlewareThrottle(s.routeDeleteAdminLogout))
	s.router.AddRoute("/api/v1/admin/serversettings", http.MethodGet, s.middlewareAuth(s.routeGetAdminServerSettings))
	s.router.AddRoute("/api/v1/admin/serversettings", http.MethodPost, s.middlewareAuth(s.routePostAdminServerSettings))
	s.router.AddRoute("/yeet", http.MethodGet, http.HandlerFunc(s.routeGetYeeted))
}

func (s *Service) routeGetMultiplayerServers(w http.ResponseWriter, r *http.Request) {
	localCacheData, _ := s.cache[LocalMultiplayer].(*CacheResponse)

	cacheData, ok := s.cache[Multiplayer].(*CacheResponse)
	if !ok {
		// if we don't have something in the cache, populate it.
		cacheData, localCacheData = s.maintenanceMultiplayerServersCache()
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Last-Modified", cacheData.Time.Format(time.RFC1123))

	if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == service.LocalhostAddress {
		_, _ = w.Write(localCacheData.Response)
		return
	}

	_, _ = w.Write(cacheData.Response)
}

func (s *Service) routeGetYeeted(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Location", "https://youtu.be/pY725Ya74VU")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
