package main

import (
	"context"
	"fmt"
	"github.com/buildboxapp/app/pkg/config"
	"github.com/labstack/gommon/color"
	"github.com/restream/reindexer"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"

	"github.com/buildboxapp/lib"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/lib/metric"

	"github.com/urfave/cli"

	stdlog "github.com/labstack/gommon/log"

	"io"
)

const sep = string(os.PathSeparator)

var fileLog *os.File
var outpurLog io.Writer


func main()  {
	warning := color.Red("[Fail]")

	// закрываем файл с логами
	defer fileLog.Close()

	defaultConfig, err := lib.DefaultConfig()
	if err != nil {
		return
	}
	rootDir, err := lib.RootDir()
	if err != nil {
		return
	}

	appCLI := cli.NewApp()
	appCLI.Usage = "Demon Buildbox Proxy started"
	appCLI.Commands = []cli.Command{
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
					Value:	rootDir,
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// инициируем пакеты
	var cfg = config.New()
	cfg.Load(configfile)

	///////////////// ЛОГИРОВАНИЕ //////////////////
	// формирование пути к лог-файлам и метрикам
	if cfg.LogsDir == "" {
		cfg.LogsDir = "logs"
	}
	// если путь указан относительно / значит задан абсолютный путь, иначе в директории
	if cfg.LogsDir[:1] != sep {
		rootDir, _ := lib.RootDir()
		cfg.LogsDir = rootDir + sep + "upload" + sep + cfg.Domain + sep + cfg.LogsDir
	}

	// инициализировать лог и его ротацию
	var logger = log.New(
		cfg.LogsDir,
		cfg.LogsLevel,
		lib.UUID(),
		cfg.Domain,
		"gui",
		cfg.UidGui,
		cfg.LogIntervalReload.Value,
		cfg.LogIntervalClearFiles.Value,
		cfg.LogPeriodSaveFiles,
	)
	logger.RotateInit(ctx)

	fmt.Printf("\n%s Enabled logs. Level:%s, Dir:%s\n", done, cfg.LogsLevel, cfg.LogsDir)
	logger.Info("Запускаем gui-сервис: ",cfg.Domain)

	// создаем метрики
	metrics := metric.New(ctx, logger, cfg.LogIntervalMetric.Value)

	defer func() {
		rec := recover()
		if rec != nil {
			b := string(debug.Stack())
			logger.Panic(fmt.Errorf("%s", b), "Recover panic from main function.")
			os.Exit(1)
		}
	}()







	///////////////// ЛОГИРОВАНИЕ //////////////////
	// кладем в глабольные переменные
	cfg.Domain 		= Config["domain"]
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


	// задаем глобальную переменную BuildBox (от нее строятся пути при загрузке шаблонов)
	app.ServiceMetrics = ServiceMetrics
	app.State = Config //state
	app.State["Workingdir"] = dir

	// для завершения сервиса ждем сигнал в процесс
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Kill)
	go ListenForShutdown(ch)


	// инициализируем кеширование
	app.State["Namespace"] 	= Replace(app.State["Domain"], "/", "_", -1)
	app.State["UrlProxy"]	= app.State["AddressProxyPointsrc"]
	app.Logger = log

	// включено кеширование
	if app.State["CachePointsrc"] != "" {
		app.DB = reindexer.NewReindex(app.State["CachePointsrc"])
		err := app.DB.OpenNamespace(app.State["Namespace"], reindexer.DefaultNamespaceOptions(), ValueCache{})
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

	if app.Get("Domain") != "" {
		app.State["ClientPath"] = "/" + app.Get("Domain") + "/ru"
	}

	//fmt.Println(app.State["client_path"] )

	var dirTemplate = app.State["Workingdir"] + "/gui/templates/*.html"
	fmt.Printf("%s Load template directory: %s\n", done, dirTemplate)
	log.Info("Load template directory: ", dirTemplate)

	router := NewRouter(ServiceMetrics) //.StrictSlash(true)

	router.Use(ServiceMetrics.Middleware)

	//router.Use(AuthProcessor)
	router.Use(Recover)

	router.PathPrefix("/upload/").Handler(http.StripPrefix("/upload/", http.FileServer(http.Dir(app.State["Workingdir"] + "/upload"))))
	router.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", http.FileServer(http.Dir(app.State["Workingdir"] + "/templates"))))

	fmt.Printf("%s Starting APP-service: %s\n", done, app.Get("PortAPP"))
	log.Info("Starting APP-service: ", app.Get("PortAPP"))

	stdlog.Fatal(http.ListenAndServe(":"+app.Get("PortAPP"), router))
}