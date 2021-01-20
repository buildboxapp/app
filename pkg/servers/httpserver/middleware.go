package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/lib"
	bbmetric "github.com/buildboxapp/lib/metric"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"
)

func (h *httpserver) JsonHeaders(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		handler.ServeHTTP(w, r)
	})
}

func (h *httpserver) MiddleLogger(inner http.Handler, name string, serviceMetrics bbmetric.ServiceMetric) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		signPage := false

		url_signin := h.cfg.UrlSingin

		u, _ := url.Parse(url_signin)
		if u.Host == r.Host && strings.Contains(u.Path, r.URL.Path) && r.URL.Path != "/" {
			// завершаем предыдущую сессию
			app.Logger.Info("Logout: ", r.URL.Path)

			Logout(w, r)
			signPage = true
		}

		// ФИКС
		// если куку взяь нельзя, а получили токен, значит используем его при авторизации
		// токен - это также кука, но переданная в урле и получанная ранее при формировани запроса
		cookie := GetCookie(w, r, "sessionID")
		token := ""
		if cookie == "" {
			token = r.FormValue("token")
		}

		authData, ok := Authorization(w, r, token)
		if !signPage {
			// получил контекст запроса + проверил авторизацию пользователя
			if !ok && name != "LogRead" && name != "TriggerMapReloadHTTP" && name != "TriggerMapHTTP" && name != "TriggerRunHTTP" && name != "PLogOut" && name != "LicenseActivate" && name != "PLogin" && name != "CLoginAutentification" && name != "Automator" && name != "JSQuery" && name != "CObjPost" && name != "CToolsFixData2" && name != "GPayYandexСonfirmation" && name != "loadToReindexer"  && name != "ProxyPing"   {
				//r.Header.Set("Referer", ClientPath + r.RequestURI) // оно не используется, но пусть будет на всякий случай
				if !strings.Contains(h.cfg.ClientPath+r.RequestURI, "page/registration") && !strings.Contains(h.cfg.ClientPath+r.RequestURI, "page/forgot") {
					if url_signin != "" {
						http.Redirect(w, r, url_signin+"?ref="+h.cfg.ClientPath+r.RequestURI, 302)
					} else {
						http.Redirect(w, r, h.cfg.ClientPath+"/login?ref="+h.cfg.ClientPath+r.RequestURI, 302)
					}
					return
				}
			}
		}

		// если без авторизации, но надо открыть страницу авторизации - делаем пустой объект, чтобы не было ошибок при генерации
		if authData == nil {
			var authData = seance.ProfileData{}
			authData.CountLicense = 5
		}

		bb, _ := json.Marshal(authData)
		ctx0 := context.WithValue(r.Context(), "User", authData)
		ctx := context.WithValue(ctx0, "UserRaw", string(bb))

		inner.ServeHTTP(w, r.WithContext(ctx))
		timeInterval := time.Since(start)
		if name != "ProxyPing" {
			h.logger.Info(
				r.Method,"; ",
				r.RequestURI,"; ",
				name,"; ",
				timeInterval,
			)
		}
		// сохраняем статистику всех запросов, в том числе и пинга (потому что этот зарпос фиксируется в количестве)
		serviceMetrics.SetTimeRequest(timeInterval)
	})
}

func (h *httpserver) MiddleAuthProcessor(next http.Handler) http.Handler {
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
		if strings.TrimSpace(authKey) != h.cfg.UidGui && r.URL.Path != "/ping" {
			lib.ResponseJSON(w, nil, "Unauthorized", nil, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *httpserver) MiddleRecover(next http.Handler) http.Handler {
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