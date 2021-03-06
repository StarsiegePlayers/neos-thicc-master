package httpd

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"

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

const (
	DayDuration = time.Hour * 24
)

type HTTPAdminSettings struct {
	*config.Configuration
	LogList map[service.ID]service.Info
	HTTPError
}

type HTTPAdminPowerAction struct {
	Action string
	HTTPError
}

func (s *Service) routeGetAdminLogin(w http.ResponseWriter, r *http.Request) {
	errorUnauthorized := &HTTPAdminLogin{
		LoggedIn: false,
		Version:  s.services.Config.BuildInfo.Version,
		Expiry:   time.Now(),

		HTTPError: HTTPError{
			Error:     "error retrieving login status",
			ErrorCode: http.StatusUnauthorized,
		},
	}

	cache, ok := s.cache[cacheAdminSessions].(map[string]*HTTPAdminSession)
	if !ok {
		s.logs.HTTPD.LogAlertf("error retrieving admin session cache")
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
		_, ok := s.services.Config.Values.HTTPD.Admins[session.Username]
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
			Version:  s.services.Config.BuildInfo.Version,
			Expiry:   session.LoginTime,
		})

		return
	}

	// user is not logged in
	s.router.jsonOut(w, HTTPAdminLogin{
		LoggedIn: false,
		Version:  s.services.Config.BuildInfo.Version,
		Expiry:   time.Time{},
	})
}

func (s *Service) routePostAdminLogin(w http.ResponseWriter, r *http.Request) {
	errorInvalidUsernameOrPassword := &HTTPAdminLogin{
		LoggedIn: false,
		Version:  s.services.Config.BuildInfo.Version,
		Expiry:   time.Now(),
		HTTPError: HTTPError{
			Error:     "invalid username or password",
			ErrorCode: http.StatusUnauthorized,
		},
	}

	errorInvalidEntry := &HTTPAdminLogin{
		LoggedIn: false,
		Version:  s.services.Config.BuildInfo.Version,
		Expiry:   time.Now(),

		HTTPError: HTTPError{
			Error:     "invalid JSON provided",
			ErrorCode: http.StatusUnprocessableEntity,
		},
	}

	errorLoggingIn := &HTTPAdminLogin{
		LoggedIn: false,
		Version:  s.services.Config.BuildInfo.Version,
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

	hash, ok := s.services.Config.Values.HTTPD.Admins[form.Username]
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
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	cookie = &http.Cookie{
		Name:     "refresh",
		Value:    td.Refresh.Token,
		Expires:  time.Now().Add(DayDuration),
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	output := &HTTPAdminLogin{
		LoggedIn: true,
		Username: form.Username,
		Version:  s.services.Config.BuildInfo.Version,
		Expiry:   time.Now().Add(time.Hour),
	}
	s.router.jsonOut(w, output)
}

func (s *Service) routeDeleteAdminLogout(w http.ResponseWriter, r *http.Request) {
	cache, ok := s.cache[cacheAdminSessions].(map[string]*HTTPAdminSession)
	if !ok {
		s.logs.HTTPD.LogAlertf("error retrieving admin session cache")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	uid, err := s.adminExtractTokenData(r)
	if err != nil {
		uid = ""

		s.logs.HTTPD.LogAlertf("error: %s", err)
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
			Version:  s.services.Config.BuildInfo.Version,
			Expiry:   time.Time{},
		})

		return
	}

	s.router.jsonOut(w, HTTPAdminLogin{
		LoggedIn: false,
		Version:  s.services.Config.BuildInfo.Version,
		Expiry:   time.Now(),

		HTTPError: HTTPError{
			Error:     "not logged in",
			ErrorCode: http.StatusUnprocessableEntity,
		},
	})
}

func (s *Service) routeGetAdminServerSettings(w http.ResponseWriter, _ *http.Request) {
	s.router.jsonOut(w, HTTPAdminSettings{
		Configuration: s.services.Config.Values,
		LogList:       service.List,
		HTTPError:     HTTPError{},
	})
}

func (s *Service) routeGetAdminServiceStatus(w http.ResponseWriter, _ *http.Request) {
	output := make(map[string]string)

	for k, v := range *s.services.Map {
		output[k.String()] = v.Status().String()
	}

	s.router.jsonOut(w, output)
}

func (s *Service) routePostAdminServerSettings(w http.ResponseWriter, r *http.Request) {
	decode := json.NewDecoder(r.Body)
	form := &HTTPAdminSettings{}
	err := decode.Decode(form)

	if err != nil {
		s.router.jsonOut(w, HTTPError{
			Error:     "invalid JSON provided",
			ErrorCode: http.StatusUnprocessableEntity,
		})

		return
	}

	s.services.Config.Values = form.Configuration
	err = s.services.Config.Write()

	if err != nil {
		form.Error = "error while writing config file to disk"
		form.ErrorCode = 1001

		s.logs.HTTPD.LogAlertf("error while writing config file to disk [%w]", err)
	}

	go s.services.Config.Rehash()

	s.router.jsonOut(w, form)
}

func (s *Service) routePostAdminPowerAction(w http.ResponseWriter, r *http.Request) {
	decode := json.NewDecoder(r.Body)
	form := &HTTPAdminPowerAction{}

	err := decode.Decode(form)
	if err != nil {
		s.router.jsonOut(w, HTTPError{
			Error:     "invalid JSON provided",
			ErrorCode: http.StatusUnprocessableEntity,
		})

		return
	}

	output := HTTPAdminPowerAction{
		Action: form.Action,
		HTTPError: HTTPError{
			Error:     "",
			ErrorCode: 0,
		},
	}

	switch form.Action {
	case "shutdown":
		s.router.jsonOut(w, output)

		go s.services.Config.Callback.Shutdown()

	case "restart":
		s.router.jsonOut(w, output)

		go s.services.Config.Callback.Restart()

	default:
		s.router.jsonOut(w, HTTPError{
			Error:     "unknown action requested",
			ErrorCode: http.StatusBadRequest,
		})
	}
}
