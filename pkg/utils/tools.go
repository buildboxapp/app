package utils

import (
	"fmt"
	"github.com/buildboxapp/app/pkg/config"
	bblib "github.com/buildboxapp/lib"
	"github.com/buildboxapp/lib/log"
)

type utils struct {
	cfg config.Config
}

type Utils interface {
	AddressProxy()
}

func (u *utils) SetAddressProxy()  {
	// если автоматическая настройка портов
	if u.cfg.AddressProxyPointsrc != "" && u.cfg.PortAutoInterval != "" {
		var portDataAPI bblib.Response
		// запрашиваем порт у указанного прокси-сервера
		u.cfg.UrlProxy = u.cfg.AddressProxyPointsrc + "port?interval=" + u.cfg.PortAutoInterval

		app.Curl("GET", u.cfg.UrlProxy, "", &portDataAPI)
		app.State["Portapp"] = fmt.Sprint(portDataAPI.Data)

		app.Logger.Info("Get: ", u.cfg.UrlProxy, "; Get PortAPP: ", u.cfg.PortApp)
	}

	// если порт передан явно через консоль, то запускаем на этом порту
	if port != "" {
		u.cfg.PortApp = port
	}

	if app.State["Portapp"] == "" {
		fmt.Print(fail, " Port APP-service is null. Servive not running.\n")
		app.Logger.Panic(err, "Port APP-service is null. Servive not running.")
	}
	log.Warning("From "+u.cfg.UrlProxy+" get PortAPP:", u.cfg.PortApp, " Domain:", u.cfg.Domain)

}

func New(cfg config.Config) Utils {
	return &utils{
		cfg,
	}
}