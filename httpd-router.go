package main

import (
	"net/http"
	"strings"
)

type Router struct {
	mux *http.ServeMux
	// routes["route"]["method"]
	routes  map[string]map[string]http.HandlerFunc
	apiPath string
	Component
}

type RouteLogger struct {
	http.ResponseWriter
	Status int
}

func (r RouteLogger) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func NewHttpRouter(apiPath string) (out *Router) {
	out = &Router{
		mux:     http.NewServeMux(),
		routes:  make(map[string]map[string]http.HandlerFunc),
		apiPath: apiPath,
		Component: Component{
			Name:   "HTTPD Router",
			LogTag: "httpd-router",
		},
	}
	return
}

func (rt *Router) Mux() (out *http.ServeMux) {
	return rt.mux
}

func (rt *Router) AddRoute(path string, method string, fn http.HandlerFunc) {
	if rt.routes[path] == nil {
		rt.routes[path] = make(map[string]http.HandlerFunc)
		rt.mux.HandleFunc(path, rt.log(rt.router))
	}
	rt.routes[path][method] = fn
}

func (rt *Router) log(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := RouteLogger{
			ResponseWriter: w,
			Status:         http.StatusOK,
		}
		fn(logger, r)
		rt.Log("[%s] %s %s - %d", r.RemoteAddr, r.Method, r.RequestURI, logger.Status)
	}
}

func (rt *Router) router(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "CERN/2.15")
	w.Header().Add("X-DummyThiccMasterVersion", VERSION)
	w.Header().Add("X-DummyThiccMeme", "https://youtu.be/pY725Ya74VU")

	if group, ok := rt.routes[r.RequestURI]; ok {
		if fn, ok := group[r.Method]; ok {
			fn(w, r)
			return
		}
		rt.errorHandler(http.StatusMethodNotAllowed, w, r)
		return
	}
	rt.errorHandler(http.StatusNotFound, w, r)
	return
}

func (rt *Router) errorHandler(code int, w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.RequestURI, rt.apiPath) {
		if routes, ok := rt.routes["default"]; ok {
			if fn, ok := routes["get"]; ok {
				fn(w, r)
			}
		}
	}
	w.WriteHeader(code)
	_, _ = w.Write([]byte(http.StatusText(code)))
}
