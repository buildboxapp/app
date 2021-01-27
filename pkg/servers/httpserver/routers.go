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

	rt := new(mux.Router)
	rt.HandleFunc("/alive", handler.Alive).Methods("GET")
	rt.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	router.Use(h.Recover)
	router.Use(h.metric.Middleware)

	router.PathPrefix("/upload/").Handler(http.StripPrefix("/upload/", http.FileServer(http.Dir(h.cfg.Workingdir+"/upload/"))))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(h.cfg.Workingdir+"/upload/gui/static/"))))
	router.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(h.cfg.Workingdir+"/templates/"))))
	router.PathPrefix("/assets/gui/templates/").Handler(http.StripPrefix("/assets/gui/templates/", http.FileServer(http.Dir(h.cfg.Workingdir+"/templates/"))))
	router.PathPrefix("/assets/gui/static/").Handler(http.StripPrefix("/assets/gui/static/", http.FileServer(http.Dir(h.cfg.Workingdir+"/upload/gui/static/"))))

	//api
	apiRouter := rt.PathPrefix("/gui/v1").Subrouter()
	apiRouter.Use(h.JsonHeaders)

	apiRouter.HandleFunc("/ping", handler.Ping).Methods("GET")

	var routes = Routes{

		// запросы (настроенные)
		Route{"ProxyPing", "GET", "/ping",  handler.Ping},

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
		handler = h.MiddleLogger(handler, route.Name, h.metric)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return rt
}
