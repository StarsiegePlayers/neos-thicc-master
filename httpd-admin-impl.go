package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type HTTPAdminCache struct {
	UUID      string
	Username  string
	LoginTime time.Time
}

func (s *HTTPDService) routeGetAdminLoginStatus(w http.ResponseWriter, r *http.Request) {
	jsonOut, err := json.Marshal(struct {
		LoggedIn bool
		User     string
	}{
		LoggedIn: false,
		User:     "",
	})
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

func (s *HTTPDService) routePostAdminLogin(w http.ResponseWriter, r *http.Request) {

}

func (s *HTTPDService) routePutAdminServerSettings(w http.ResponseWriter, r *http.Request) {

}

func (s *HTTPDService) routeGetAdminServerSettings(w http.ResponseWriter, r *http.Request) {

}
