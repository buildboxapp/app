// запускаем указанные виды из поддерживаемых серверов
package servers

import (
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/app/pkg/servers/httpserver"
	"github.com/buildboxapp/app/pkg/service"
	bbmetric "github.com/buildboxapp/lib/metric"
	"strings"
)

type servers struct {
	mode       string
	service    service.Service
	httpserver httpserver.Server
	metrics    bbmetric.ServiceMetric
	cfg        model.Config
}

type Servers interface {
	Run()
}

// запускаем указанные севрера
func (s *servers) Run() {
	if strings.Contains(s.mode, "http") {
		s.httpserver.Run()
	}
}

func New(
	mode string,
	service service.Service,
	httpserver httpserver.Server,
	metrics bbmetric.ServiceMetric,
	cfg model.Config,
) Servers {
	return &servers{
		mode,
		service,
		httpserver,
		metrics,
		cfg,
	}
}