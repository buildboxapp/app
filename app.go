package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/labstack/gommon/color"
	"github.com/restream/reindexer"

	"os/signal"

	bblib "github.com/buildboxapp/lib"
	"github.com/urfave/cli"
	. "github.com/buildboxapp/app/lib"
	"github.com/buildboxapp/logger"

	stdlog "github.com/labstack/gommon/log"
	buildboxapp"github.com/buildboxapp/app/lib"

	"io"
)

var fileLog *os.File
var outpurLog io.Writer

var log = logger.Log{}
var lib = bblib.Lib{}
var app = buildboxapp.App{}


func init() {

	app.Init()

	// задаем настройки логирования выполнения функций библиотеки
	lib.Logger = &log
	app.Logger = &log
}

func main()  {

	// закрываем файл с логами
	defer fileLog.Close()

	defaultConfig, err := lib.DefaultConfig()
	if err != nil {
		log.Warning("Warning! The default configuration directory was not found.")
	}


	appCLI := cli.NewApp()
	appCLI.Usage = "Demon Buildbox Proxy started"
	appCLI.Commands = []cli.Command{
		{
			Name:"run",
			Usage: "Run demon Buildbox APP process",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:	"config, c",
					Usage:	"Название файла конфигурации, с которым будет запущен сервис",
					Value:	defaultConfig,
				},
				&cli.StringFlag{
					Name:	"dir, d",
					Usage:	"Путь к шаблонам",
					Value:	lib.RootDir(),
				},
				&cli.StringFlag{
					Name:	"port, p",
					Usage:	"Порт, на котором запустить процесс",
					Value:	"",
				},
			},
			Action: func(c *cli.Context) error {
				configfile := c.String("config")
				dir := c.String("dir")
				lib.RunProcess(configfile, dir, "app", "start","services")

				return nil
			},
		},
		{
			Name:"start", ShortName: "",
			Usage: "Start single Buildbox APP process",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:	"config, c",
					Usage:	"Название файла конфигурации, с которым будет запущен сервис",
					Value:	defaultConfig,
				},
				cli.StringFlag{
					Name:	"dir, d",
					Usage:	"Путь к шаблонам",
					Value:	lib.RootDir(),
				},
				cli.StringFlag{
					Name:	"port, p",
					Usage:	"Порт, на котором запустить процесс",
					Value:	"",
				},
			},
			Action: func(c *cli.Context) error {
				configfile := c.String("config")
				port := c.String("port")
				dir := c.String("dir")

				Start(configfile, dir, port)

				return nil
			},
		},
	}

	appCLI.Run(os.Args)

	return
}

// стартуем сервис приложения
func Start(configfile, dir, port string) {

	//for k, v := range FuncMap {
	//	log.Info(k, " = ", v, " conf: ", configfile)
	//}

	Config, _, err := lib.ReadConf(configfile)
	if err != nil {
		log.Error(err)
	}

	///////////////// ЛОГИРОВАНИЕ //////////////////
	// кладем в глабольные переменные
	Domain 		= Config["domain"]
	LogsDir 	= Config["gui_logs"]
	LogsLevel 	= Config["gui_level_logs_pointsrc"]
	// формирование пути к лог-файлам и метрикам
	if LogsDir == "" {
		LogsDir = "logs"
	}
	// если путь указан относительно / значит задан абсолютный путь, иначе в директории
	if LogsDir[:1] != "/" {
		LogsDir = lib.RootDir() + "/upload/" + Domain + "/" + LogsDir
	}
	log.Init(LogsDir, LogsLevel, UUID(), Domain, "app")
	log.Info("Запускаем app-сервис: ",Domain)
	//////////////////////////////////////////////////


	done := color.Green("OK")
	fail := color.Red("FAIL")

	state, _, err := lib.ReadConf(configfile)
	if err != nil {
		log.Fatal(err)
	}

	// задаем глобальную переменную BuildBox (от нее строятся пути при загрузке шаблонов)
	app.State = state
	app.State["workdir"] = dir

	// для завершения сервиса ждем сигнал в процесс
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Kill)
	go ListenForShutdown(ch)

	// индексируем шаблоны, если режим - production (единоразовая обработка шаблонов)
	//if !debugMode {
	//	dirTemplate := lib.CurrentDir() + "/upload/control/unify.bm/template/*.html" //Application["workingdir"] + "/upload/control/unify.bm/template/*.html"
	//	t = template.Must(template.New("").Funcs(FuncMap).ParseGlob(dirTemplate))
	//}

	proxy_url := ""

	log.UID = UidPrecess
	log.Name = app.Get("domain")
	log.Service = "app"


	// если автоматическая настройка портов
	if app.Get("address_proxy_pointsrc") != "" && app.Get("port_auto_interval") != "" {
		var portDataAPI bblib.Response
		// запрашиваем порт у указанного прокси-сервера
		proxy_url = app.Get("address_proxy_pointsrc") + "port?interval=" + app.Get("port_auto_interval")

		app.Curl("GET", proxy_url, "", &portDataAPI)
		app.State["PortAPP"] = fmt.Sprint(portDataAPI.Data)

		app.Logger.Info("Get: ", proxy_url, "; Get PortAPP: ", app.State["PortAPP"])
	}

	// если порт передан явно через консоль, то запускаем на этом порту
	if port != "" {
		app.State["PortAPP"] = port
	}

	if app.State["PortAPP"] == "" {
		fmt.Print(fail, " Port APP-service is null. Servive not running.\n")
		app.Logger.Fatal(err, "Port APP-service is null. Servive not running.")
	}
	log.Warning("From "+proxy_url+" get PortAPP:", app.Get("PortAPP"), " Domain:", app.Get("domain"))

	// инициализируем кеширование
	app.State["namespace"] 	= Replace(app.Get("domain"), "/", "_", -1)
	app.State["url_proxy"]	= app.Get("address_proxy_pointsrc")

	// включено кеширование
	if app.Get("cache_pointsrc") != "" {
		app.DB = reindexer.NewReindex(app.Get("cache_pointsrc"))
		err := app.DB.OpenNamespace(app.Get("namespace"), reindexer.DefaultNamespaceOptions(), ValueCache{})
		if err != nil {
			fmt.Printf("%s Error connecting to database. Plaese check this parameter in the configuration. %s\n", fail, app.Get("cache_pointsrc"))
			fmt.Printf("%s\n", err)
			app.Logger.Error(err, "Error connecting to database. Plaese check this parameter in the configuration: ", app.Get("cache_pointsrc"))
			return
		} else {
			fmt.Printf("%s Cache-service is running", done)
			app.Logger.Info("Cache-service is running")
			app.State["BaseCache"] = "on"
		}
	}

	if app.Get("domain") != "" {
		app.State["client_path"] = "/" + app.Get("domain") + "/ru"
	}

	//fmt.Println(app.State["client_path"] )

	var dirTemplate = app.State["workdir"] + "/gui/templates/*.html"
	fmt.Printf("\n%s Load template directory: %s\n", done, dirTemplate)
	log.Info("Load template directory: ", dirTemplate)

	router := NewRouter() //.StrictSlash(true)
	router.PathPrefix("/upload/").Handler(http.StripPrefix("/upload/", http.FileServer(http.Dir(app.State["workdir"] + "/upload"))))
	router.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(app.State["workdir"] + "/templates"))))

	fmt.Printf("%s Starting APP-service: %s\n", done, app.Get("PortAPP"))
	log.Info("Starting APP-service: ", app.Get("PortAPP"))

	stdlog.Fatal(http.ListenAndServe(":"+app.Get("PortAPP"), router))
}
