package httpd

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/StarsiegePlayers/neos-thicc-master/src/config"
	"github.com/StarsiegePlayers/neos-thicc-master/src/log"
	"github.com/StarsiegePlayers/neos-thicc-master/src/service"
)

type Router struct {
	mux *http.ServeMux

	// routes["route"]["method"]
	routes map[string]map[string]http.Handler

	emedFS    http.FileSystem
	buildInfo *service.BuildInfo

	*log.Log
}

type HTTPError struct {
	Error     string
	ErrorCode int
}

type RouteLogger struct {
	http.ResponseWriter
	Status int
}

func (r *RouteLogger) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func NewHTTPRouter(log *log.Log, buildinfo *service.BuildInfo) (out *Router) {
	out = &Router{
		mux:       http.NewServeMux(),
		routes:    make(map[string]map[string]http.Handler),
		Log:       log,
		buildInfo: buildinfo,
	}
	out.mux.HandleFunc("/", out.log(out.router))

	return
}

func (rt *Router) Mux() (out *http.ServeMux) {
	return rt.mux
}

func (rt *Router) AddRoute(path string, method string, fn http.Handler) {
	if _, ok := rt.routes[path]; !ok {
		rt.routes[path] = make(map[string]http.Handler)
	}

	rt.routes[path][method] = fn
}

func (rt *Router) SetFileSystem(fs fs.FS, err error) {
	if err != nil {
		rt.Log.LogAlertf("error parsing embedded filesystem?")
	}

	rt.emedFS = http.FS(fs)
}

func (rt *Router) jsonOut(w http.ResponseWriter, arg interface{}) {
	output, err := json.Marshal(arg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(output)
}

func (rt *Router) log(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := RouteLogger{
			ResponseWriter: w,
			Status:         http.StatusOK,
		}
		fn(&logger, r)
		rt.Log.Logf("[%s] %s %s - %d", r.RemoteAddr, r.Method, r.RequestURI, logger.Status)
	}
}

func (rt *Router) router(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "CERN/2.15")
	w.Header().Add("X-DummyThiccMasterVersion", rt.buildInfo.Version)
	w.Header().Add("X-DummyThiccMeme", config.EggURL)

	// first match on the api overlay
	if group, ok := rt.routes[r.RequestURI]; ok {
		if fn, ok := group[r.Method]; ok {
			fn.ServeHTTP(w, r)
			return
		}

		rt.errorHandler(http.StatusMethodNotAllowed, w, r)

		return
	}

	// then match on the emedded filesystem
	if r.Method == http.MethodGet {
		x, err := rt.emedFS.Open(r.RequestURI)
		if err == nil {
			_ = x.Close()

			rt.serveFile(w, r)

			return
		}
	}

	// if the url does not contain the api path, serve up the index
	if !strings.HasPrefix(strings.ToLower(r.RequestURI), "/api") {
		index, err := rt.emedFS.Open("index.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fi, err := index.Stat()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, fi.Name(), fi.ModTime(), index)

		return
	}

	// if still can't find something, 404 out
	rt.errorHandler(http.StatusNotFound, w, r)
}

func (rt *Router) errorHandler(code int, w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(code)
	_, _ = w.Write([]byte(http.StatusText(code)))
}

func (rt *Router) serveFile(w http.ResponseWriter, r *http.Request) {
	http.FileServer(rt.emedFS).ServeHTTP(w, r)
}
