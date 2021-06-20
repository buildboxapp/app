package main

import (
	"context"
	"fmt"
	"github.com/buildboxapp/app/pkg/cache"
	"github.com/buildboxapp/app/pkg/function"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/app/pkg/servers"
	"github.com/buildboxapp/app/pkg/servers/httpserver"
	"github.com/buildboxapp/app/pkg/service"
	"github.com/buildboxapp/app/pkg/utils"
	"github.com/labstack/gommon/color"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/buildboxapp/lib"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/lib/config"
	"github.com/buildboxapp/lib/metric"

)

const sep = string(os.PathSeparator)

func main()  {
	lib.RunServiceFuncCLI(Start)
}

// стартуем сервис приложения
func Start(configfile, dir, port, mode string) {
	var cfg model.Config
	done := color.Green("[OK]")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// инициируем пакеты
	err := config.Load(configfile, &cfg)
	if err != nil {
		fmt.Printf("%s (%s)", "Error. Load config is failed.", err)
		return
	}

	cfg.UidService = strings.Split(configfile, ".")[0]

	// формируем значение переменных по-умолчанию или исходя из требований сервиса
	cfg.SetClientPath()
	cfg.SetRootDir()
	cfg.SetConfigName()

	cfg.Workingdir = cfg.RootDir

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
	// инициализируем кеширование
	cfg.Namespace	= strings.ReplaceAll(cfg.Domain, "/", "_")
	cfg.UrlProxy	= cfg.AddressProxyPointsrc

	// инициализировать лог и его ротацию
	var logger = log.New(
		cfg.LogsDir,
		cfg.LogsLevel,
		lib.UUID(),
		cfg.Domain,
		"app",
		cfg.UidService,
		cfg.LogIntervalReload.Value,
		cfg.LogIntervalClearFiles.Value,
		cfg.LogPeriodSaveFiles,
	)
	logger.RotateInit(ctx)

	fmt.Printf("%s Enabled logs. Level:%s, Dir:%s\n", done, cfg.LogsLevel, cfg.LogsDir)
	logger.Info("Запускаем app-сервис: ",cfg.Domain)

	// создаем метрики
	metrics := metric.New(
		ctx,
		logger,
		cfg.LogIntervalMetric.Value,
	)

	defer func() {
		rec := recover()
		if rec != nil {
			b := string(debug.Stack())
			logger.Panic(fmt.Errorf("%s", b), "Recover panic from main function.")
			cancel()
			os.Exit(1)
		}
	}()

	ult := utils.New(
		cfg,
		logger,
	)

	fnc := function.New(
		cfg,
		ult,
		logger,
	)


	port = ult.AddressProxy()
	cfg.PortApp = port

	cach := cache.New(
		cfg,
		logger,
		fnc,
		ult,
	)

	// собираем сервис
	src := service.New(
		logger,
		cfg,
		metrics,
		ult,
		cach,
	)

	// httpserver
	httpserver := httpserver.New(
		ctx,
		cfg,
		src,
		metrics,
		logger,
	)

	// для завершения сервиса ждем сигнал в процесс
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	go ListenForShutdown(ch, logger)

	srv := servers.New(
		"http",
		src,
		httpserver,
		metrics,
		cfg,
	)
	srv.Run()

}

func ListenForShutdown(ch <- chan os.Signal, logger log.Log)  {
	var done = color.Grey("[OK]")

	<- ch
	logger.Warning("Service is stopped. Logfile is closed.")
	logger.Close()
	fmt.Printf("%s Service is stopped. Logfile is closed.\n", done)
	os.Exit(0)
}
