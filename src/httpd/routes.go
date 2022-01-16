package httpd

import (
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/stun"
)

const HTTPStatusEnhanceYourCalm = 420

func (s *Service) registerRoutes() {
	s.router.SetFileSystem(fs.Sub(s.services.Config.BuildInfo.EmbedFS, "www-build"))
	s.router.AddRoute("/api/v1/master/info", http.MethodGet, http.HandlerFunc(s.routeGetMasterInfo))
	s.router.AddRoute("/api/v1/multiplayer/servers", http.MethodGet, http.HandlerFunc(s.routeGetMultiplayerServers))
	s.router.AddRoute("/api/v1/admin/login", http.MethodGet, s.middlewareThrottle(s.routeGetAdminLogin))
	s.router.AddRoute("/api/v1/admin/login", http.MethodPost, s.middlewareThrottle(s.routePostAdminLogin))
	s.router.AddRoute("/api/v1/admin/login", http.MethodDelete, s.middlewareThrottle(s.routeDeleteAdminLogout))
	s.router.AddRoute("/api/v1/admin/serversettings", http.MethodGet, s.middlewareAuth(s.routeGetAdminServerSettings))
	s.router.AddRoute("/api/v1/admin/serversettings", http.MethodPost, s.middlewareAuth(s.routePostAdminServerSettings))
	s.router.AddRoute("/api/v1/admin/poweraction", http.MethodPost, s.middlewareAuth(s.routePostAdminPowerAction))
	s.router.AddRoute("/yeet", http.MethodGet, http.HandlerFunc(s.routeGetYeeted))
}

func (s *Service) routeGetMultiplayerServers(w http.ResponseWriter, r *http.Request) {
	cacheData := s.cache[cacheMultiplayer].(map[string]*CacheResponse)

	data, ok := cacheData[""]
	if !ok {
		// if we don't have something in the cache, populate it.
		data = s.maintenanceMultiplayerServersCache()
	}

	remoteIPString, _, err := net.SplitHostPort(r.RemoteAddr)

	// skip if STUN service isn't running
	if s.services.STUN != nil && err == nil {
		remoteIP := net.ParseIP(remoteIPString)
		for _, v := range s.services.STUN.(*stun.Service).LocalAddresses {
			if v.Contains(remoteIP) {
				if d2, ok := cacheData[v.String()]; ok {
					data = d2
					break
				}
			}
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Last-Modified", data.Time.Format(time.RFC1123))

	_, _ = w.Write(data.Response)
}

func (s *Service) routeGetMasterInfo(w http.ResponseWriter, r *http.Request) {
	hostname := s.services.Config.Values.Service.Hostname
	if hostname == "" {
		hostname = "(no-name)"
	}

	requestHost, _, _ := net.SplitHostPort(r.RemoteAddr)
	hostname = strings.ReplaceAll(hostname, "\\n", "")

	s.router.jsonOut(w, struct {
		Hostname string
		MOTD     string
		ID       uint16
		Uptime   time.Time
	}{
		Hostname: hostname,
		MOTD:     s.services.Template.Get(requestHost),
		ID:       s.services.Config.Values.Service.ID,
		Uptime:   s.services.Config.Startup,
	})
}

func (s *Service) routeGetYeeted(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Location", config.EggURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
