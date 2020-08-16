package app_lib

import (
	"fmt"
	"net/http"
	"time"
)

func (c *App) LoggerHTTP(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		defer func() {
			if r := recover(); r != nil {
				c.Logger.Error(fmt.Errorf("%s", r), "Error. Fail generate page")
				return
			}
		}()

		inner.ServeHTTP(w, r)

		if name != "ProxyPing" && false == true {
			c.Logger.Info(
				"Query: %s %s %s %s",
				r.Method,
				r.RequestURI,
				name,
				time.Since(start),
			)
		}
	})
}
