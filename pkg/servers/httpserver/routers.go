package httpserver

import (
	"github.com/buildboxapp/app/pkg/servers/httpserver/handlers"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"

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
	rt.HandleFunc("/alive", handlers.Alive).Methods("GET")
	rt.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	router.Use(h.MiddleRecover)
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
		Route{"PIndex", "GET", "/", handler.PIndex},
		Route{"PIndex", "POST", "/role/{role}", handler.PIndex},
		Route{"ProxyPing", "GET", "/ping",  handler.Ping},

		// запросы (настроенные)
		Route{"JSQuery", "GET", "/query/{obj}", handler.JSQuery},
		Route{"JSQuery", "POST", "/query/{obj}", handler.JSQuery},
		Route{"JSQuery", "OPTION", "/query/{obj}", handler.JSQuery},

		Route{"JSQuery", "GET", "/query/{obj}/{option}", handler.JSQuery},
		Route{"JSQuery", "POST", "/query/{obj}/{option}", handler.JSQuery},
		Route{"JSQuery", "OPTION", "/query/{obj}/{option}", handler.JSQuery},

		// режим конструктора
		Route{"PColumnTab", "GET", "/list/{view}/tab/{obj}", handler.MColumnTab},
		Route{"PColumnTab", "GET", "/list/{view}/content/{obj}", handler.MColumnContent},
		Route{"PColumnTab", "GET", "/list/{view}/full/{obj}", handler.MColumnFull},


		// list - отображение внутри системы
		Route{"GList", "GET", "/list/{view}/{obj}", handler.GList},
		Route{"GList", "POST", "/list/{view}/{obj}", handler.GList},
		Route{"GList", "GET", "/list/{view}/{obj}/{option}", handler.GList},
		Route{"GList", "POST", "/list/{view}/{obj}/{option}", handler.GList},

		// view - внешнее отображение в стилистике консоли
		Route{"GView", "GET", "/view/{view}/{obj}", handler.GView},
		Route{"GView", "POST", "/view/{view}/{obj}", handler.GView},
		Route{"GView", "GET", "/view/{view}/{obj}/{option}", handler.GView},
		Route{"GView", "POST", "/view/{view}/{obj}/{option}", handler.GView},

		// modal - внутреннее только HTML (без внешних стилей)
		Route{"GModal", "GET", "/modal/{view}/{obj}", handler.GModal},
		Route{"GModal", "POST", "/modal/{view}/{obj}", handler.GModal},
		Route{"GModal", "GET", "/modal/{view}/{obj}/{option}", handler.GModal},
		Route{"GModal", "POST", "/modal/{view}/{obj}/{option}", handler.GModal},

		// json/pdf/doc/text - внутреннее только в форматах (с загрузкой и просмотром)
		Route{"GJson", "GET", "/json/{view}/{obj}/{option}", handler.GJson},
		Route{"GJson", "GET", "/jsonload/{view}/{obj}/{option}", handler.GJsonLoad},
		Route{"GJson", "GET", "/pdf/{view}/{obj}/{option}", handler.GPdf},
		Route{"GJson", "GET", "/pdfload/{view}/{obj}/{option}", handler.GPdfLoad},
		Route{"GJson", "GET", "/doc/{view}/{obj}/{option}", handler.GDoc},
		Route{"GJson", "GET", "/docload/{view}/{obj}/{option}", handler.GDocLoad},


		// form - внешнее отображение форм ввода данных

		Route{"CSelect", "POST", "/select", handler.CSelect},

		// создание объекта в запросе. можем загружать как из:
		// POST формы - отправлены поля (по-умолчанию, без доп. параметров)
		// JSON-формат (&format=json) - получаем в теле поля, но в JSON-не (надо разобрать по-другому)
		// FILE (&file=путь_к_файлу_от_корня) - загружаем из файла готовый объект
		Route{"CObjPost", "POST", "/objs", handler.CObjPost},

		Route{"CLoginAutentification", "GET", "/login/auth", handler.CLoginAutentification},
		Route{"CLoginAutentification", "POST", "/login/auth", handler.CLoginAutentification},
		Route{"PLogin", "GET", "/login", handler.PLogin},
		Route{"PLogOut", "GET", "/logout", handler.PLogout},

		// запуск триггера
		Route{"TriggerRunHTTP", "POST", "/trigger/{id}", Thandler.riggerRunHTTP},
		Route{"TriggerRunHTTP", "GET", "/trigger/{id}", handler.TriggerRunHTTP},
		Route{"TriggerMapReloadHTTP", "POST", "/triggers/reload", handler.TriggerMapReloadHTTP},
		Route{"TriggerMapReloadHTTP", "GET", "/triggers/reload", handler.TriggerMapReloadHTTP},
		Route{"TriggerMapHTTP", "GET", "/triggers/map", handler.TriggerMapHTTP},

		// отдельно получаем компоненты интерфейса
		Route{"CNavigator", "GET", "/component/navigator/{id}", handler.CNavigator},
		Route{"CNavigator", "POST", "/component/navigator/_change/{role}", handler.CNavigator}, // не используется отдельно без перезагрузки дашборда

		// формат SIMPLE OBJECT - как отображаем (form, modal) / чем открываем (шаблон) / что открываем (объект)
		// создаем объект (по-умолчанию тип Данные)
		Route{"GForm", "GET", "/obj/{view}/_tpls/{tpl}/_source/{source}", handler.GForm},

		Route{"GForm", "GET", "/obj/{view}/{obj}", handler.GForm},
		Route{"GForm", "POST", "/obj/{view}/{obj}", handler.GForm},
		Route{"GForm", "GET", "/obj/{view}/{obj}/{tpl}", handler.GForm},
		Route{"GForm", "POST", "/obj/{view}/{obj}/{tpl}", handler.GForm},
		Route{"GForm", "GET", "/obj/{view}/{obj}/{tpl}/{source}", handler.GForm},
		Route{"GForm", "POST", "/obj/{view}/{obj}/{tpl}/{source}", handler.GForm},

		// изменяем объект (объект не указан)

		// запрос на блокировку поля при редактировании (проверка по ревизии)
		Route{"CElementBlock", "POST", "/element/block", handler.CElementBlock},
		Route{"CElementUpdate", "GET", "/element/update", handler.CElementUpdate},
		Route{"CElementUpdate", "POST", "/element/update", handler.CElementUpdate},
		Route{"CElementCheckup", "POST", "/element/checkup", handler.CElementCheckup},

		// работа с линками
		Route{"CElementBlock", "GET", "/link/{obj}", handler.CLinkGet},
		Route{"CElementUpdate", "POST", "/link/add", handler.CLinkAdd},
		Route{"CElementCheckup", "POST", "/link/delete", handler.CLinkDelete},

		// обработка запуска/остановки сервисов/проектов
		Route{"ServiceActionHTTP", "GET", "/service", handler.ServiceActionHTTP},

		Route{"CToolsLoadfile", "POST", "/tools/loadfile", handler.CToolsLoadfile},
		Route{"CToolsCreatefile", "POST", "/tools/createfile", handler.CToolsCreatefile},

		Route{"CToolsTexteditor", "GET", "/tools/texteditor", handler.CToolsTexteditor},
		Route{"CToolsTexteditor", "POST", "/tools/texteditor", handler.CToolsTexteditor},
		Route{"CToolsTplTree", "GET", "/tools/tpltree", handler.CToolsTplTree},


		// платежный модуль Yandex.Money
		Route{"GPayYandex", "GET", "/tools/pay", handler.GPayYandex},
		// адрес для уведомлений необходимо прописывать для каждого подключаемого сайта на стороне яндекс.касса
		// в https://kassa.yandex.ru/my/shop-settings -> URL для уведомлений
		// например: https://buildbox.app/buildbox/gui/tools/pay/confirmation
		Route{"GPayYandexСonfirmation", "POST", "/tools/pay/confirmation", handler.GPayYandexСonfirmation},


		// вызов ручной манипуляции данных (код может менятся для конкретной задачи)
		//Route{"CToolsFixData2", "GET", "/tools/fixdata", CToolsFixData2},
		Route{"CToolsUpdateLinks", "POST", "/tools/updatelinks", handler.CToolsUpdateLinks},
		Route{"CToolsUpdateLinks", "POST", "/tools/updatetitles", handler.CToolsUpdateTitles},

		// дублируем объекты/шаблоны
		// перечень uid-ов объектов передается в параметре ?objs=....,....,.... (ч/з запятую)
		Route{"CToolsDubl", "POST", "/tools/dubl", handler.CToolsDubl},

		Route{"loadToReindexer", "GET", "/tools/loadreindex", handler.loadToReindexer},


		Route{"PTrash", "POST", "/trash/{option}", handler.ETrash},
		// можно передавать несколкьо uid-ов через ,
		Route{"PTrash", "POST", "/trash/{option}/{uids}", handler.ETrash},
		Route{"Automator", "POST", "/automator", handler.Automator},

		// лицензирование (активация)
		Route{"LicensePush", "GET", "/license/push/{obj}", handler.LicensePush},
		Route{"LicenseActivate", "GET", "/license/activate/{obj}", handler.LicenseActivate},

		// чтение лог.файлов
		Route{"LogRead", "GET", "/tools/logs", handler.LogRead},

		// импорт/экспорт объектов
		Route{"Export", "POST", "/tools/export/{option}", handler.Export},
		Route{"Export", "GET", "/tools/export/{option}", handler.Export},
		Route{"Import", "POST", "/tools/import", handler.Import},

		//Route{"CUpdateStart", "POST", "/update/start", CUpdateStart},

		// GUI - запросы для реализации логики консоли v1.0 через стандартный APP
		Route{"ENavigatorHTTP", "GET", "/ui/navigator", handler.ENavigatorHTTP},
		Route{"EFormHTTP", "GET", "/ui/form", handler.EFormHTTP},

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
