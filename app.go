package main

import (
	"context"
	"fmt"
	"github.com/labstack/gommon/color"
	"github.com/restream/reindexer"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"time"

	. "github.com/buildboxapp/app/lib"
	bblib "github.com/buildboxapp/lib"
	bblog "github.com/buildboxapp/lib/log"
	bbmetric "github.com/buildboxapp/lib/metric"

	"github.com/urfave/cli"

	buildboxapp "github.com/buildboxapp/app/lib"
	stdlog "github.com/labstack/gommon/log"

	"io"
)

var fileLog *os.File
var outpurLog io.Writer

var log = &bblog.Log{}
var lib = bblib.Lib{}
var app = buildboxapp.App{}

var logIntervalReload = 10 * time.Minute			// интервал проверки необходимости пересозданния нового файла
var logIntervalClearFiles = 30 * time.Minute		// интервал проверка на необходимость очистки старых логов
var logPeriodSaveFiles = "0-1-0"				// период хранения логов
var logIntervalMetric = 10 * time.Second			// период сохранения метрик в файл логирования

var ServiceMetrics bbmetric.ServiceMetric

func init() {

	app.Init()

	// задаем настройки логирования выполнения функций библиотеки
	app.Logger = log
}

func main()  {

	// закрываем файл с логами
	defer fileLog.Close()

	defaultConfig, err := lib.DefaultConfig()
	if err != nil {
		log.Warning("Warning! The default configuration directory was not found.")
	}

	defer func() {
		if r := recover(); r != nil {
			log.Warning(fmt.Errorf("%s", r), "Error. Fail generate page")
			return
		}
	}()

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
	done := color.Green("[OK]")
	fail := color.Red("[FAIL]")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func() {
		rec := recover()
		if rec != nil {
			b := string(debug.Stack())
			log.Panic(fmt.Errorf("%s", b), "Recover panic from main function.")
		}
	}()

	Config, _, err := lib.ReadConf(configfile)
	if err != nil {
		log.Error(err)
	}

	///////////////// ЛОГИРОВАНИЕ //////////////////
	// кладем в глабольные переменные
	Domain 		= Config["domain"]
	LogsDir 	= Config["app_logs"]
	LogsLevel 	= Config["app_level_logs_pointsrc"]
	UidAPP		= Config["data-uid"]
	ReplicasService, err = strconv.Atoi(Config["replicas_app"])
	if err != nil {
		ReplicasService = 1
	}

	// формирование пути к лог-файлам и метрикам
	if LogsDir == "" {
		LogsDir = "logs"
	}
	// если путь указан относительно / значит задан абсолютный путь, иначе в директории
	if LogsDir[:1] != "/" {
		LogsDir = lib.RootDir() + "/upload/" + Domain + "/" + LogsDir
	}

	fmt.Printf("%s Enabled logs. Level:%s, Dir:%s\n", done, LogsLevel, LogsDir)

	// инициализировать лог и его ротацию
	log = bblog.New(LogsDir, LogsLevel, bblib.UUID(), Domain, "app", UidAPP, logIntervalReload, logIntervalClearFiles, logPeriodSaveFiles)
	log.RotateInit(ctx)

	lib.Logger = log	// инициируем логер в либе

	log.Info("Запускаем app-сервис: ",Domain)
	//////////////////////////////////////////////////

	// создаем метрики
	ServiceMetrics = bbmetric.New(ctx, log, logIntervalMetric)

	state, _, err := lib.ReadConf(configfile)
	if err != nil {
		log.Panic(err)
	}

	// задаем глобальную переменную BuildBox (от нее строятся пути при загрузке шаблонов)
	app.ServiceMetrics = ServiceMetrics
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
		app.Logger.Panic(err, "Port APP-service is null. Servive not running.")
	}
	log.Warning("From "+proxy_url+" get PortAPP:", app.Get("PortAPP"), " Domain:", app.Get("domain"))

	// инициализируем кеширование
	app.State["namespace"] 	= Replace(app.Get("domain"), "/", "_", -1)
	app.State["url_proxy"]	= app.Get("address_proxy_pointsrc")
	app.Logger = log

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
	fmt.Printf("%s Load template directory: %s\n", done, dirTemplate)
	log.Info("Load template directory: ", dirTemplate)

	router := NewRouter(ServiceMetrics) //.StrictSlash(true)

	router.Use(ServiceMetrics.Middleware)

	//router.Use(AuthProcessor)
	router.Use(Recover)

	router.PathPrefix("/upload/").Handler(http.StripPrefix("/upload/", http.FileServer(http.Dir(app.State["workdir"] + "/upload"))))
	router.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(app.State["workdir"] + "/templates"))))

	fmt.Printf("%s Starting APP-service: %s\n", done, app.Get("PortAPP"))
	log.Info("Starting APP-service: ", app.Get("PortAPP"))

	stdlog.Fatal(http.ListenAndServe(":"+app.Get("PortAPP"), router))
}
