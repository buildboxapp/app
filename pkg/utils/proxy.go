package utils

import (
	"fmt"
	bblib "github.com/buildboxapp/lib"
	"github.com/labstack/gommon/color"
)


func (u *utils) AddressProxy()()  {
	fail := color.Red("[Fail]")

	// если автоматическая настройка портов
	if u.cfg.AddressProxyPointsrc != "" && u.cfg.PortAutoInterval != "" {
		var portDataAPI bblib.Response
		// запрашиваем порт у указанного прокси-сервера
		u.cfg.UrlProxy = u.cfg.AddressProxyPointsrc + "port?interval=" + u.cfg.PortAutoInterval

		u.Curl("GET", u.cfg.UrlProxy, "", &portDataAPI)
		u.cfg.PortApp = fmt.Sprint(portDataAPI.Data)

		u.logger.Info("Get: ", u.cfg.UrlProxy, "; Get PortAPP: ", u.cfg.PortApp)
	}

	// если порт передан явно через консоль, то запускаем на этом порту
	//if port != "" {
	//	u.cfg.PortApp = port
	//}

	if u.cfg.PortApp == "" {
		fmt.Print(fail, " Port APP-service is null. Servive not running.\n")
		u.logger.Panic(nil, "Port APP-service is null. Servive not running.")
	}
	u.logger.Warning("From "+u.cfg.UrlProxy+" get PortAPP:", u.cfg.PortApp, " Domain:", u.cfg.Domain)

}