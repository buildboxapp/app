package handlers

import (
	"fmt"
	"github.com/buildboxapp/app/pkg/service"
	"net/http"
)

// запуск сервиса с GUI (прнимаем id-объекта, сохраняем в файл-конфигурации и дергаем api)
func ServiceActionHTTP(w http.ResponseWriter, r *http.Request) {
	var result string

	//d := Authorization(w, r)

	// для операций запуска сервиса
	action 	:= r.FormValue("action")
	config 	:= r.FormValue("config")
	mode 	:= r.FormValue("mode")

	// для операций остановки и перезагрузки не всего проекта, а только части сервисов
	domain := r.FormValue("domain")
	alias := r.FormValue("alias")

	r.ParseForm()
	body1 := r.PostForm

	//fmt.Println("body1: ",body1)

	switch action {
	case "start", "stop", "reload":
		result = service.ServiceAction(action, config, domain, alias)
	case "build":	// для сборки нужны параметры и контекст (данные о пользователе)
		result = service.ServiceBuild(r, config, body1)
	case "baseactivate":	// активация переданной базы данных
		result = service.ServiceBaseActivate(body1)
	case "saveconfig":
		err := service.ServiceSaveConfig(config, mode)
		if err != nil {
			result = fmt.Sprint(err)
		} else {
			result = "OK! Сonfiguration file created."
		}
	}

	w.Write([]byte(result))
}
