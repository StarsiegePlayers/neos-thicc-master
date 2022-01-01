package httpd

import (
	"net"
	"net/http"
)

func (s *Service) middlewareAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := s.adminIsTokenValid(r)
		if err != nil {
			// no cookie / logged out? 401
			s.router.jsonOut(w, HTTPError{
				Error:     "unauthorized",
				ErrorCode: http.StatusUnauthorized,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) middlewareThrottle(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		throttle := s.cache[Throttle].(map[string]int)
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if throttle[host] > s.Config.Values.HTTPD.MaxRequestsPerMinute {
			s.router.jsonOut(w, HTTPError{
				Error:     "enhance your calm",
				ErrorCode: HTTPStatusEnhanceYourCalm,
			})
			return
		}
		s.Lock()
		throttle[host]++
		s.Unlock()
		next.ServeHTTP(w, r)
	})
}
