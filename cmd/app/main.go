package main

import (
	"context"
	"fmt"
	"github.com/buildboxapp/app/pkg/cache"
	"github.com/buildboxapp/app/pkg/config"
	"github.com/buildboxapp/app/pkg/function"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/labstack/gommon/color"
	"github.com/restream/reindexer"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"

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

	fmt.Printf("\n%s Enabled logs. Level:%s, Dir:%s\n", done, cfg.LogsLevel, cfg.LogsDir)
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

	// для завершения сервиса ждем сигнал в процесс
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Kill)
	go ListenForShutdown(ch)

	fnc := function.New(
		cfg,
		*logger,
	)

	cach := cache.New(
		cfg,
		*logger,
		fnc,
	)

}



func ListenForShutdown(ch <- chan os.Signal)  {
	<- ch
	os.Exit(0)
}
