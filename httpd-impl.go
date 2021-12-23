package main

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"time"

	"github.com/StarsiegePlayers/darkstar-query-go/v2/query"
)

type MasterQuery struct {
	*query.MasterQuery
	ServerCount int
}

type ServerListData struct {
	RequestTime time.Time
	Masters     []*MasterQuery
	Games       []*query.PingInfoQuery
	Errors      []string
}

func (s *HTTPDService) registerRoutes() {
	s.router.SetFileSystem(fs.Sub(wwwFS, "www-build"))
	s.router.AddRoute("/api/v1/admin/serversettings", http.MethodGet, s.middlewareAuth(s.routeGetAdminServerSettings))
	s.router.AddRoute("/api/v1/admin/serversettings", http.MethodPut, s.middlewareAuth(s.routePutAdminServerSettings))
	s.router.AddRoute("/api/v1/admin/login", http.MethodPost, s.middlewareThrottle(s.routePostAdminLogin))
	s.router.AddRoute("/api/v1/admin/login", http.MethodGet, s.middlewareThrottle(s.routeGetAdminLoginStatus))
	s.router.AddRoute("/api/v1/multiplayer/servers", http.MethodGet, http.HandlerFunc(s.routeGetMultiplayerServers))
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
	data, ok := s.cache["multiplayer"].(*ServerListData)

	if !ok || data.RequestTime.Before(time.Now().Add(-60*time.Second)) {
		mstr := (*s.Services)["master"].(*MasterService)
		mstr.Lock()
		games := make([]*query.PingInfoQuery, 0)
		for _, v := range mstr.ServerList {
			games = append(games, v.PingInfoQuery)
		}
		mstr.Unlock()

		poll := (*s.Services)["poll"].(*PollService)
		poll.Lock()
		errors := make([]string, 0)
		for _, v := range poll.PollMasterInfo.Errors {
			errors = append(errors, v.Error())
		}

		masters := make([]*MasterQuery, 0)
		for _, v := range poll.PollMasterInfo.Masters {
			masters = append(masters, &MasterQuery{
				MasterQuery: v,
				ServerCount: len(v.Servers),
			})
		}
		poll.Unlock()

		data = &ServerListData{
			RequestTime: time.Now(),
			Masters:     masters,
			Games:       games,
			Errors:      errors,
		}
		s.cache["multiplayer"] = data
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Last-Modified", data.RequestTime.Format(time.RFC1123))

	jsonOut, err := json.Marshal(data)
	if err != nil {
		s.LogAlert("error marshalling api server list %s", err)
		return
	}

	_, err = w.Write(jsonOut)
	if err != nil {
		s.LogAlert("error writing api server list %s", err)
		return
	}
}

func (s *HTTPDService) routeGetAdminLoginStatus(w http.ResponseWriter, r *http.Request) {

}

func (s *HTTPDService) routePostAdminLogin(w http.ResponseWriter, r *http.Request) {

}

func (s *HTTPDService) routePutAdminServerSettings(w http.ResponseWriter, r *http.Request) {

}

func (s *HTTPDService) routeGetAdminServerSettings(w http.ResponseWriter, r *http.Request) {

}
