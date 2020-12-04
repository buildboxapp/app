package main

import (
	"fmt"
	stdlog "log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
	. "github.com/buildboxapp/app/lib"

)

func Logger(next http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		if name != "ProxyPing"  { //&& false == true
			mes := fmt.Sprintf("Query: %s %s %s %s",
				r.Method,
				r.RequestURI,
				name,
				time.Since(start))
			stdlog.Printf(mes)
			log.Info(mes)
		}
	})
}

func AuthProcessor(next http.Handler) http.Handler {
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
		if strings.TrimSpace(authKey) != UidAPP && r.URL.Path != "/ping" {
			lib.ResponseJSON(w, nil, "Unauthorized", nil, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(r *http.Request) {
			rec := recover()
			if rec != nil {
				b := string(debug.Stack())
				fmt.Println(r.URL.String())
				log.Panic(fmt.Errorf("%s", b), "Recover panic from path: ", r.URL.String(), "; form: ", r.Form)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}(r)
		next.ServeHTTP(w, r)
	})
}