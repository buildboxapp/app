package service

import (
	"context"
	"github.com/buildboxapp/app/pkg/block"
	"github.com/buildboxapp/app/pkg/cache"
	"github.com/buildboxapp/app/pkg/config"
	"github.com/buildboxapp/app/pkg/function"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/app/pkg/utils"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/lib/metric"
)

type service struct {
	logger log.Log
	cfg config.Config
	metrics metric.ServiceMetric
	utils utils.Utils
	cache cache.Cache
	block block.Block
	function function.Function
}

// Service interface
type Service interface {
	Ping(ctx context.Context) (result []model.Pong, err error)
	Page(ctx context.Context, in model.ServiceIn) (out model.ServicePageOut, err error)
	Block(ctx context.Context, in model.ServiceIn) (out model.ServiceBlockOut, err error)
}

func New(
	logger log.Log,
	cfg config.Config,
	metrics metric.ServiceMetric,
	utils utils.Utils,
	cache cache.Cache,
) Service {
	var tplfunc = function.NewTplFunc(cfg, logger)
	var function = function.New(cfg, logger)
	var blocks = block.New(cfg, logger, utils, function, tplfunc)

	return &service{
		logger,
		cfg,
		metrics,
		utils,
		cache,
		blocks,
		function,
	}
}
