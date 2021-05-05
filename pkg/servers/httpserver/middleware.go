package httpserver

import (
	"fmt"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/lib"
	"github.com/buildboxapp/lib/log"
	bbmetric "github.com/buildboxapp/lib/metric"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

func (h *httpserver) MiddleLogger(next http.Handler, name string, logger log.Log, serviceMetrics bbmetric.ServiceMetric) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)
		timeInterval := time.Since(start)
		if name != "ProxyPing"  { //&& false == true
			mes := fmt.Sprintf("Query: %s %s %s %s",
				r.Method,
				r.RequestURI,
				name,
				timeInterval)
			logger.Info(mes)
		}

		// сохраняем статистику всех запросов, в том числе и пинга (потому что этот запрос фиксируется в количестве)
		serviceMetrics.SetTimeRequest(timeInterval)
	})
}

func (h *httpserver) AuthProcessor(next http.Handler, cfg model.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authKey string

		authKeyHeader := r.Header.Get("X-Auth-Key")
		if authKeyHeader != "" {
			authKey = authKeyHeader
		} else {
			authKeyCookie, err := r.Cookie("X-Auth-Key")
			if err == nil {
				authKey = authKeyCookie.Value
			}
		}

		// не передали ключ (пропускаем пинги)
		if strings.TrimSpace(authKey) == "" && r.URL.Path != "/ping" {
			lib.ResponseJSON(w, nil, "Unauthorized", nil, nil)
			return
		}

		// не соответствие переданного ключа и UID-а API (пропускаем пинги)
		if strings.TrimSpace(authKey) != cfg.UidService && r.URL.Path != "/ping" {
			lib.ResponseJSON(w, nil, "Unauthorized", nil, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *httpserver) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(r *http.Request) {
			rec := recover()
			if rec != nil {
				b := string(debug.Stack())
				//fmt.Println(r.URL.String())
				h.logger.Panic(fmt.Errorf("%s", b), "Recover panic from path: ", r.URL.String(), "; form: ", r.Form)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}(r)
		next.ServeHTTP(w, r)
	})
}

func (h *httpserver) JsonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}
