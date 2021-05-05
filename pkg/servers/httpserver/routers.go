package httpserver

import (
	"github.com/buildboxapp/app/pkg/servers/httpserver/handlers"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"net/http/pprof"
)

type Result struct {
	Status  string `json:"status"`
	Content []interface{}
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func (h *httpserver) NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	handler := handlers.New(h.src, h.logger, h.cfg)

	router.HandleFunc("/alive", handler.Alive).Methods("GET")
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	//apiRouter := rt.PathPrefix("/gui/v1").Subrouter()
	//router.Use(h.JsonHeaders)

	var routes = Routes{
		// запросы (настроенные)
		Route{"ProxyPing", "GET", "/ping",  handler.Ping},

		Route{"Cache", "GET", "/tools/cacheclear", handler.Cache},

		Route{"Page", "GET", "/", handler.Page},
		Route{"Page", "GET", "/{page}", handler.Page},
		Route{"Page", "POST", "/{page}", handler.Page},
		Route{"Page", "GET", "/{page}/", handler.Page},
		Route{"Page", "POST", "/{page}/", handler.Page},

		Route{"Block", "GET", "/block/{block}", handler.Block},
		Route{"Block", "POST", "/block/{block}", handler.Block},
		Route{"Block", "GET", "/block/{block}/", handler.Block},
		Route{"Block", "POST", "/block/{block}/", handler.Block},

		// Регистрация pprof-обработчиков
		Route{"pprofIndex", "GET", "/debug/pprof/", pprof.Index},
		Route{"pprofIndex", "GET", "/debug/pprof/cmdline", pprof.Cmdline},
		Route{"pprofIndex", "GET", "/debug/pprof/profile", pprof.Profile},
		Route{"pprofIndex", "GET", "/debug/pprof/symbol", pprof.Symbol},
		Route{"pprofIndex", "GET", "/debug/pprof/trace", pprof.Trace},
	}

	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = h.MiddleLogger(handler, route.Name, h.logger, h.metric)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	router.Use(h.Recover)
	router.Use(h.metric.Middleware)

	router.StrictSlash(true)

	router.PathPrefix("/upload/").Handler(http.StripPrefix("/upload/", http.FileServer(http.Dir(h.cfg.Workingdir + "/upload"))))
	router.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(h.cfg.Workingdir + "/templates"))))

	return router
}
