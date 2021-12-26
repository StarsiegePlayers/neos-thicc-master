package main

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"sort"
	"time"

	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
)

type MasterQuery struct {
	*query.MasterQuery
	ServerCount int
}

func (s *HTTPDService) registerRoutes() {
	s.router.SetFileSystem(fs.Sub(wwwFS, "www-build"))
	s.router.AddRoute("/api/v1/admin/serversettings", http.MethodGet, s.middlewareAuth(s.routeGetAdminServerSettings))
	s.router.AddRoute("/api/v1/admin/serversettings", http.MethodPut, s.middlewareAuth(s.routePutAdminServerSettings))
	s.router.AddRoute("/api/v1/admin/login", http.MethodPost, s.middlewareThrottle(s.routePostAdminLogin))
	s.router.AddRoute("/api/v1/admin/login", http.MethodGet, s.middlewareThrottle(s.routeGetAdminLoginStatus))
	s.router.AddRoute("/api/v1/multiplayer/servers", http.MethodGet, http.HandlerFunc(s.routeGetMultiplayerServers))
	s.router.AddRoute("/yeet", http.MethodGet, http.HandlerFunc(s.routeGetYeeted))
}

func (s *HTTPDService) middlewareAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Our middleware logic goes here...
		next.ServeHTTP(w, r)
	})
}

func (s *HTTPDService) middlewareThrottle(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Our middleware logic goes here...
		next.ServeHTTP(w, r)
	})
}

func (s *HTTPDService) routeGetMultiplayerServers(w http.ResponseWriter, r *http.Request) {
	cacheData, ok := s.cache[HTTPCacheMultiplayer].(*CacheResponse)
	// if we don't have something in the cache, populate it.
	if !ok {
		cacheData = s.maintenanceMultiplayerServersCache()
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Last-Modified", cacheData.Time.Format(time.RFC1123))
	w.Write(cacheData.Response)
}

func (s *HTTPDService) routeGetYeeted(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Location", "https://youtu.be/pY725Ya74VU")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *HTTPDService) maintenanceMultiplayerServersCache() (cacheData *CacheResponse) {
	games := make([]*query.PingInfoQuery, 0)
	errors := make([]string, 0)
	masters := make([]*MasterQuery, 0)

	// MasterService should never be nil, but just in case
	if s.MasterService != nil {
		s.MasterService.Lock()
		for _, v := range s.MasterService.ServerList {
			games = append(games, v.PingInfoQuery)
		}
		s.MasterService.Unlock()
	}

	// skip if poll service isn't running
	if s.PollService != nil {
		s.PollService.Lock()
		for _, v := range s.PollService.PollMasterInfo.Errors {
			errors = append(errors, v.Error())
		}

		for _, v := range s.PollService.PollMasterInfo.Masters {
			masters = append(masters, &MasterQuery{
				MasterQuery: v,
				ServerCount: len(v.Servers),
			})
		}
		s.PollService.Unlock()
	}

	// update the cache
	data := &ServerListData{
		RequestTime: time.Now(),
		Masters:     masters,
		Games:       games,
		Errors:      errors,
	}

	sort.Sort(data.Masters)
	sort.Sort(data.Games)

	jsonOut, err := json.Marshal(data)
	if err != nil {
		s.LogAlert("error marshalling api server list %s", err)
		return
	}

	cacheData = &CacheResponse{
		Response: jsonOut,
		Time:     data.RequestTime,
	}

	s.Lock()
	s.cache[HTTPCacheMultiplayer] = cacheData
	s.Unlock()
	return
}
