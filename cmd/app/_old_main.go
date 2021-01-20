package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/braintree/manners"
	"github.com/buildboxapp/app/pkg/servers"
	"github.com/buildboxapp/app/pkg/servers/httpserver"
	"github.com/buildboxapp/app/pkg/service"
	bblib "github.com/buildboxapp/lib"
	"github.com/buildboxapp/lib/log"
	bbmetric "github.com/buildboxapp/lib/metric"
	"github.com/labstack/gommon/color"
	bbcli "github.com/buildboxapp/app/pkg/cli"
	"github.com/urfave/cli"
	"html/template"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/buildboxapp/app/pkg/config"

	bbapp "github.com/buildboxapp/app/lib"
)

var t *template.Template

var app = bbapp.App{}
var TriggersTimerOn = TriggerMap{Data: map[string]Trigger{}}
var ServiceMetrics bbmetric.ServiceMetric
var sep = string(filepath.Separator)

func init()  {
	// это нужно для добавлении во внутреннние функции расширений (функции из пакета внешнего)
	app.Init()
	app.Logger = log

	// инициируем уборщика лишних сессий в заданном интервале
	go func() {
		for {
			err := sessionInMemory.SessionCollector()
			if err != nil {
				return
			}	// останавливаем горутину если прибился основной процесс и объекта сессий нет
			time.Sleep(30 * time.Second)
		}
		return
	}()
}

func main()  {
	defaultConfig, err := bblib.DefaultConfig()
	rootDir, err := bblib.RootDir()
	if err != nil {
		fmt.Println(err,"Warning! The default configuration directory was not found.")
	}

	app := cli.NewApp()
	app.Usage = "Demon Buildbox Studio started"
	app.Commands = []cli.Command{
		{
			Name:"start",
			Usage: "Start Buildbox Workplace process",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:	"config, c",
					Usage:	"Название файла конфигурации, с которым будет запущен сервис",
					Value:	defaultConfig,
				},
				&cli.StringFlag{
					Name:	"dir, d",
					Usage:	"Путь к рабочей директории приложения",
					Value:	rootDir,
				},
				&cli.StringFlag{
					Name:	"port, p",
					Usage:	"Порт на котором будет запущен сервис",
					Value:	"",
				},
			},
			Action: func(c *cli.Context) error {
				configfile := c.String("config")
				port := c.String("port")
				dir := c.String("dir")

				start(configfile, dir, port)

				return nil
			},
		},
	}

	app.Run(os.Args)

	return
}

func start(configfile, dir, port string) {
	done := color.Green("[OK]")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// инициируем пакеты
	var cfg = config.New(configfile)

	///////////////// ЛОГИРОВАНИЕ //////////////////
	// формирование пути к лог-файлам и метрикам
	if cfg.LogsDir == "" {
		cfg.LogsDir = "logs"
	}
	// если путь указан относительно / значит задан абсолютный путь, иначе в директории
	if cfg.LogsDir[:1] != sep {
		rootDir, _ := bblib.RootDir()
		cfg.LogsDir = rootDir + sep + "upload" + sep + cfg.Domain + sep + cfg.LogsDir
	}

	// инициализировать лог и его ротацию
	var logger = log.New(cfg.LogsDir, cfg.LogsLevel, bblib.UUID(), cfg.Domain, "gui", cfg.UidGui, cfg.LogIntervalReload.Value, cfg.LogIntervalClearFiles.Value, cfg.LogPeriodSaveFiles)
	logger.RotateInit(ctx)

	fmt.Printf("\n%s Enabled logs. Level:%s, Dir:%s\n", done, cfg.LogsLevel, cfg.LogsDir)
	logger.Info("Запускаем gui-сервис: ",cfg.Domain)

	// создаем метрики
	metrics := bbmetric.New(ctx, logger, cfg.LogIntervalMetric.Value)

	defer func() {
		rec := recover()
		if rec != nil {
			b := string(debug.Stack())
			logger.Panic(fmt.Errorf("%s", b), "Recover panic from main function.")
		}
	}()

	// получаю порт GUI через обращение к прокси, если указан интервал и адрес прокси-сервера
	if cfg.AddressProxy != "" && cfg.PortAutoInterval != "" {
		var portDataGUI bblib.Response
		proxy_url_port := cfg.AddressProxy + "port?interval=" + cfg.PortAutoInterval

		// запрашиваем GUI-порт у указанного прокси-сервера - первый свободный
		bblib.Curl("GET", proxy_url_port, "", &portDataGUI, map[string]string{}, cfg.UrlApi, cfg.UrlGui)
		cfg.PortGui = fmt.Sprint(portDataGUI.Data)
	}

	// бывает, если стартуется корневой прокси-сервис, для него обращаемся к localhost:порт
	if cfg.AddressProxy == "" {
		cfg.AddressProxy = "http://localhost:" + cfg.PortProxy
	}

	// защищаемся от разнонаписания адреса прокси в настройках
	lastSlesh := cfg.AddressProxy[len(cfg.AddressProxy)-1:]
	if lastSlesh == "/" {
		cfg.AddressProxy = cfg.AddressProxy[:len(cfg.AddressProxy)-1]
	}

	cfg.UrlApi = cfg.AddressProxy + "/" + cfg.Domain + "/api/v1/"
	cfg.UrlGui = cfg.AddressProxy + "/" + cfg.Domain + "/gui/"
	cfg.SetClientPath()	// задаем значнеие ClientPath

	// параметры для приема платежей (Яндекс.Касса)
	cfg.YandexRedirecturl 	= cfg.Domain + "/gui/list/page/licenses"
	cfg.Workingdir = dir

	var DirTemplate = cfg.Workingdir + "/upload/gui/templates/*.html"
	fmt.Printf("%s Load template directory: %s\n", done, DirTemplate)
	logger.Info("Load template directory (",configfile,"): ", done, "; ", DirTemplate)

	// в качестве фукнций передаем фунции описанные в APP
	t = template.Must(template.New("").Funcs(FuncMap).ParseGlob(DirTemplate))
	router := NewRouter(ServiceMetrics) //.StrictSlash(true)


	// преобразуем текущую конфигурацию в map[string]string
	b1, _ := json.Marshal(cfg)
	json.Unmarshal(b1, &Config)

	app.State = Config
	app.Logger = logger

	if cfg.Domain != "" {
		app.State["Projectuid"] = "/" + cfg.Domain + "/gui"
	}

	if cfg.CheckServiceautomator != "" {
		trigger.TriggerTimer(ctx)
	}

	fmt.Printf("%s Starting GUI-service: %s\n", done, cfg.PortGui)
	logger.Info("Starting GUI-service: (",configfile,"): OK")

	logger.Exit(manners.ListenAndServe(":"+cfg.PortGui, router))

	src := service.New(nil, *logger)
	httpsrv := httpserver.New(
		ctx,
		cfg,
		src,
		configfile,
		dir,
		port,
		metrics,
		*logger,
	)


	servers := servers.New(
		"http",
		src,
		httpsrv,
		metrics,
		cfg,
	)

	clid := bbcli.New(servers, *logger)
	clid.Run()
}