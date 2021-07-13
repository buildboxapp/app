package httpserver

import (
	"context"
	"fmt"
	"github.com/buildboxapp/app/pkg/iam"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/app/pkg/service"
	"github.com/buildboxapp/app/pkg/session"
	"github.com/buildboxapp/app/pkg/utils"
	"github.com/buildboxapp/lib/log"
	bbmetric "github.com/buildboxapp/lib/metric"
	"github.com/labstack/gommon/color"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	// should be so!
	_ "github.com/buildboxapp/app/pkg/servers/docs"
)

type httpserver struct {
	ctx     context.Context
	cfg     model.Config
	src     service.Service
	metric  bbmetric.ServiceMetric
	logger  log.Log
	utl     utils.Utils
	jtk     iam.IAM
	session session.Session
}

type Server interface {
	Run() (err error)
}

// Run server
func (h *httpserver) Run() error {
	done := color.Green("[OK]")

	// закрываем логи при завешрении работы сервера
	defer func() {
		h.logger.Warning("Service is stopped. Logfile is closed.")
		h.logger.Close()
	}()

	//err := httpscerts.Check(h.cfg.SSLCertPath, h.cfg.SSLPrivateKeyPath)
	//if err != nil {
	//	panic(err)
	//}
	srv := &http.Server{
		Addr:         ":" + h.cfg.PortApp,
		Handler:      h.NewRouter(false),	// переадресация будет работать, если сам севрис будет стартовать https-сервер (для этого надо получать сертфикаты)
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
	ctx 	context.Context,
	cfg 	model.Config,
	src 	service.Service,
	metric 	bbmetric.ServiceMetric,
	logger 	log.Log,
	utl 	utils.Utils,
	jtk 	iam.IAM,
	session session.Session,
) Server {
	return &httpserver{
		ctx,
		cfg,
		src,
		metric,
		logger,
		utl,
		jtk,
		session,
	}
}