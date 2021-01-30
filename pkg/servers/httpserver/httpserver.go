package httpserver

import (
	"context"
	"fmt"
	"github.com/buildboxapp/app/pkg/service"
	"github.com/buildboxapp/lib/log"
	bbmetric "github.com/buildboxapp/lib/metric"
	"github.com/labstack/gommon/color"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	// should be so!
	_ "github.com/buildboxapp/app/pkg/servers/docs"

	"github.com/buildboxapp/app/pkg/config"
)

type httpserver struct {
	ctx context.Context
	cfg config.Config
	src service.Service
	metric bbmetric.ServiceMetric
	logger log.Log
}

type Server interface {
	Run() (err error)
}

// Run server
func (h *httpserver) Run() error {
	done := color.Green("[OK]")

	//err := httpscerts.Check(h.cfg.SSLCertPath, h.cfg.SSLPrivateKeyPath)
	//if err != nil {
	//	panic(err)
	//}
	srv := &http.Server{
		Addr:         ":" + h.cfg.PortApp,
		Handler:      h.NewRouter(),
		ReadTimeout:  h.cfg.ReadTimeout.Value,
		WriteTimeout: h.cfg.WriteTimeout.Value,
	}
	fmt.Printf("%s Service run (port:%s)\n", done, h.cfg.PortApp)
	h.logger.Info("Запуск https сервера", zap.String("port", h.cfg.PortApp))
	//e := srv.ListenAndServeTLS(h.cfg.SSLCertPath, h.cfg.SSLPrivateKeyPath)

	e := srv.ListenAndServe()
	if e != nil {
		return errors.Wrap(e, "SERVER run")
	}
	return nil
}


func New(
	ctx context.Context,
	cfg config.Config,
	src service.Service,
	metric bbmetric.ServiceMetric,
	logger log.Log,
) Server {
	return &httpserver{
		ctx,
		cfg,
		src,
		metric,
		logger,
	}
}