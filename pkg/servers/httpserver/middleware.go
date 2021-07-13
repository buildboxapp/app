package httpserver

import (
	"fmt"
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

func (h *httpserver) AuthProcessor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authKey string
		var err error
		var XAuthToken string

		// пропускаем пинги
		if r.URL.Path == "/ping" || r.URL.Path == "/auth" || strings.Contains(r.URL.Path, "/templates") || strings.Contains(r.URL.Path, "/upload") {
			next.ServeHTTP(w, r)
			return
		}

		authKeyHeader := r.Header.Get("X-Auth-Key")
		if authKeyHeader != "" {
			authKey = authKeyHeader
		} else {
			authKeyCookie, err := r.Cookie("X-Auth-Key")
			if err == nil {
				authKey = authKeyCookie.Value
			}
		}

		// не передали ключ - вход не осуществлен. войди
		if strings.TrimSpace(authKey) == "" {
			http.Redirect(w, r, h.cfg.SigninUrl+"?ref="+h.cfg.ClientPath+r.RequestURI, 302)
			return
		}

		// валидируем токен
		status, token, refreshToken, err := h.jtk.Verify(authKey)

		// пробуем обновить пришедший токен
		if !status {
			XAuthToken, err = h.jtk.Refresh(refreshToken)

			// если токен был обновлен чуть ранее, то текущий запрос надо пропустить
			// чтобы избежать повторного обновления и дать возможность завершиться отправленным
			// единовременно нескольким запросам (как правило это интервал 5-10 секунд)
			if XAuthToken == "skip" {
				next.ServeHTTP(w, r)
				return
			}

			if err == nil && XAuthToken != "<nil>" && XAuthToken != "" {
				// заменяем куку у пользователя в браузере
				cookie := &http.Cookie{
					Path: "/",
					Name:   "X-Auth-Key",
					Value:  XAuthToken,
					MaxAge: 30000,
				}

				// после обновления получаем текущий токен
				status, token, _, err = h.jtk.Verify(XAuthToken)

				// переписываем куку у клиента
				http.SetCookie(w, cookie)
			}
		}

		// выкидываем если обновление невозможно
		if !status || err != nil {
			http.Redirect(w, r, h.cfg.SigninUrl+"?ref="+h.cfg.ClientPath+r.RequestURI, 302)
			return
		}

		// добавляем значение токена в локальный реестр сесссий (ПЕРЕДЕЛАТЬ)
		if token != nil {
			h.session.Set(token)
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

func (h *httpserver) HttpsOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// remove/add not default ports from req.Host
		target := "https://" + req.Host + req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			target += "?" + req.URL.RawQuery
		}
		// see comments below and consider the codes 308, 302, or 301
		http.Redirect(w, req, target, http.StatusTemporaryRedirect)
	})
}

