package jwtoken

import (
	"github.com/buildboxapp/app/pkg/i18n"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/app/pkg/utils"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/lib/metric"
)

type jwtoken struct {
	logger log.Log
	utils utils.Utils
	cfg model.Config
	metric metric.ServiceMetric
	msg  	i18n.I18n
}

type JWToken interface {
	Refresh(token string) (result string, err error)
	Verify(tokenString string) (statue bool, body *model.Token, refreshToken string, err error)
}

func New(logger log.Log, utils utils.Utils, cfg model.Config, metric metric.ServiceMetric, msg i18n.I18n) JWToken {
	return &jwtoken{
		logger,
		utils,
		cfg,
		metric,
		msg,
	}
}
