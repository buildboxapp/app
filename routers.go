package main

import (
	bbmetric "github.com/buildboxapp/lib/metric"
	"github.com/gorilla/mux"
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

func NewRouter(serviceMetrics bbmetric.ServiceMetric) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = MiddleLogger(handler, route.Name, serviceMetrics)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

var routes = Routes{
	Route{"PIndex", "GET", "/", app.PIndex},
	Route{"ProxyPing", "GET", "/ping",  app.ProxyPing},

	Route{"PIndex", "GET", "/{page}", app.PIndex},
	Route{"PIndex", "POST", "/{page}", app.PIndex},
	Route{"GetBlock", "GET", "/block/{block}", app.GetBlock},
	Route{"GetBlock", "POST", "/block/{block}", app.GetBlock},

	Route{"PIndex", "GET", "/{page}/", app.PIndex},
	Route{"PIndex", "POST", "/{page}/", app.PIndex},
	Route{"GetBlock", "GET", "/block/{block}/", app.GetBlock},
	Route{"GetBlock", "POST", "/block/{block}/", app.GetBlock},

	// Регистрация pprof-обработчиков
	Route{"Index", "GET", "/debug/pprof/", pprof.Index},
	Route{"Index", "GET", "/debug/pprof/cmdline", pprof.Cmdline},
	Route{"Index", "GET", "/debug/pprof/profile", pprof.Profile},
	Route{"Index", "GET", "/debug/pprof/symbol", pprof.Symbol},
	Route{"Index", "GET", "/debug/pprof/trace", pprof.Trace},

}
