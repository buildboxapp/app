package service

import (
	"context"
	"github.com/buildboxapp/app/pkg/api"
	"github.com/buildboxapp/app/pkg/block"
	"github.com/buildboxapp/app/pkg/cache"
	"github.com/buildboxapp/app/pkg/function"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/app/pkg/i18n"
	"github.com/buildboxapp/app/pkg/session"
	"github.com/buildboxapp/app/pkg/utils"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/lib/metric"
)

type service struct {
	logger   log.Log
	cfg      model.Config
	metrics  metric.ServiceMetric
	utils    utils.Utils
	cache    cache.Cache
	block    block.Block
	function function.Function
	msg 	 i18n.I18n
	session  session.Session
	api 	api.Api
}

// Service interface
type Service interface {
	Alive(ctx context.Context) (out model.AliveOut, err error)
	Ping(ctx context.Context) (result []model.Pong, err error)
	Page(ctx context.Context, in model.ServiceIn) (out model.ServicePageOut, err error)
	Block(ctx context.Context, in model.ServiceIn) (out model.ServiceBlockOut, err error)
	Cache(ctx context.Context, in model.ServiceCacheIn) (out model.RestStatus, err error)
}

func New(
	logger log.Log,
	cfg model.Config,
	metrics metric.ServiceMetric,
	utils utils.Utils,
	cache cache.Cache,
	msg i18n.I18n,
	session session.Session,
	api api.Api,
) Service {
	var tplfunc = function.NewTplFunc(cfg, utils, logger, msg, api)
	var function = function.New(cfg, utils, logger, msg, api)
	var blocks = block.New(cfg, logger, utils, function, tplfunc, api)

	return &service{
		logger,
		cfg,
		metrics,
		utils,
		cache,
		blocks,
		function,
		msg,
		session,
		api,
	}
}
