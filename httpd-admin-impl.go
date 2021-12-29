package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aykevl/pwhash"
)

type HTTPAdminLogin struct {
	LoggedIn bool
	Username string
	Password string
	Version  string
	Expiry   time.Time

	HTTPError
}

type HTTPAdminSettings struct {
	Configuration

	HTTPError
}

func (s *HTTPDService) routeGetAdminLoginStatus(w http.ResponseWriter, r *http.Request) {
	errorUnauthorized := &HTTPAdminLogin{
		LoggedIn: false,
		Version:  buildVersion,
		Expiry:   time.Now(),

		HTTPError: HTTPError{
			Error:     "error retrieving login status",
			ErrorCode: http.StatusUnauthorized,
		},
	}

	cache, ok := s.cache[HTTPCacheAdminSessions].(map[string]*HTTPAdminSession)
	if !ok {
		s.LogAlert("error retrieving admin session cache")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uid, err := s.adminExtractTokenData(r)
	if err != nil {
		uid = ""
	}

	var session *HTTPAdminSession
	if uid != "" {
		if sesh, ok := cache[uid]; ok {
			if sesh.LoginTime.Add(time.Hour).After(time.Now()) {
				// we have a valid session, update last usage
				sesh.LoginTime = time.Now()
				session = sesh
			}
		}
	}

	if session != nil {
		_, ok := s.Config.Values.HTTPD.Admins[session.Username]
		if !ok {
			// something weird happened here, log that user out
			delete(cache, uid)
			w.Header().Add("Location", "/admin")
			s.router.jsonOut(w, errorUnauthorized)
			return
		}

		// user is logged in
		s.router.jsonOut(w, HTTPAdminLogin{
			LoggedIn: true,
			Username: session.Username,
			Version:  buildVersion,
			Expiry:   session.LoginTime,
		})
		return
	}

	// user is not logged in
	s.router.jsonOut(w, HTTPAdminLogin{
		LoggedIn: false,
		Version:  buildVersion,
		Expiry:   time.Time{},
	})
	return
}

func (s *HTTPDService) routePostAdminLogin(w http.ResponseWriter, r *http.Request) {
	errorInvalidUsernameOrPassword := &HTTPAdminLogin{
		LoggedIn: false,
		Version:  buildVersion,
		Expiry:   time.Now(),
		HTTPError: HTTPError{
			Error:     "invalid username or password",
			ErrorCode: http.StatusUnauthorized,
		},
	}

	errorInvalidEntry := &HTTPAdminLogin{
		LoggedIn: false,
		Version:  buildVersion,
		Expiry:   time.Now(),

		HTTPError: HTTPError{
			Error:     "invalid JSON provided",
			ErrorCode: http.StatusUnprocessableEntity,
		},
	}

	errorLoggingIn := &HTTPAdminLogin{
		LoggedIn: false,
		Version:  buildVersion,
		Expiry:   time.Now(),
		HTTPError: HTTPError{
			Error:     "error logging in",
			ErrorCode: http.StatusUnauthorized,
		},
	}

	decode := json.NewDecoder(r.Body)
	form := &HTTPAdminLogin{}
	err := decode.Decode(form)

	if err != nil {
		s.router.jsonOut(w, errorInvalidEntry)
		return
	}

	if form.Username == "" || form.Password == "" {
		s.router.jsonOut(w, errorInvalidUsernameOrPassword)
		return
	}

	hash, ok := s.Config.Values.HTTPD.Admins[form.Username]
	if !ok {
		s.router.jsonOut(w, errorInvalidUsernameOrPassword)
		return
	}

	if !pwhash.Verify(form.Password, hash) {
		s.router.jsonOut(w, errorInvalidUsernameOrPassword)
		return
	}

	td, err := s.adminCreateToken()
	if err != nil {
		s.router.jsonOut(w, errorLoggingIn)
		return
	}

	s.Lock()
	_ = s.adminCreateSession(form.Username, td)
	s.Unlock()

	cookie := &http.Cookie{
		Name:     "token",
		Value:    td.Access.Token,
		Expires:  time.Now().Add(time.Hour * 1),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	cookie = &http.Cookie{
		Name:     "refresh",
		Value:    td.Refresh.Token,
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	output := &HTTPAdminLogin{
		LoggedIn: true,
		Username: form.Username,
		Version:  buildVersion,
		Expiry:   time.Now().Add(time.Hour * 1),
	}
	s.router.jsonOut(w, output)
}

func (s *HTTPDService) routeAdminLogout(w http.ResponseWriter, r *http.Request) {
	cache, ok := s.cache[HTTPCacheAdminSessions].(map[string]*HTTPAdminSession)
	if !ok {
		s.LogAlert("error retrieving admin session cache")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	uid, err := s.adminExtractTokenData(r)
	if err != nil {
		uid = ""
		s.LogAlert("error: %s", err)
	}

	var session *HTTPAdminSession
	if uid != "" {
		if sesh, ok := cache[uid]; ok {
			if sesh.LoginTime.Add(time.Hour).After(time.Now()) {
				// we have a valid session
				session = sesh
			}
		}
	}
	if session != nil {
		for k, v := range cache {
			if v.Username == session.Username {
				delete(cache, k)
			}
		}

		s.router.jsonOut(w, HTTPAdminLogin{
			LoggedIn: false,
			Version:  buildVersion,
			Expiry:   time.Time{},
		})
		return
	}

	s.router.jsonOut(w, HTTPAdminLogin{
		LoggedIn: false,
		Version:  buildVersion,
		Expiry:   time.Now(),

		HTTPError: HTTPError{
			Error:     "not logged in",
			ErrorCode: http.StatusUnprocessableEntity,
		},
	})

}

func (s *HTTPDService) routePutAdminServerSettings(w http.ResponseWriter, r *http.Request) {
	decode := json.NewDecoder(r.Body)
	form := &HTTPAdminSettings{}
	err := decode.Decode(form)

	if err != nil {
		s.router.jsonOut(w, HTTPAdminSettings{
			HTTPError: HTTPError{
				Error:     "invalid JSON provided",
				ErrorCode: http.StatusUnprocessableEntity,
			},
		})
		return
	}

	data, err := json.Marshal(form)

	s.LogAlert("%s %s", string(data), err)

}

func (s *HTTPDService) routeGetAdminServerSettings(w http.ResponseWriter, r *http.Request) {
	s.router.jsonOut(w, s.Config.Values)
	return
}
