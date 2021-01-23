package service

import (
	"context"
	"github.com/buildboxapp/app/pkg/config"
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
}

// Service interface
type Service interface {
	Ping(ctx context.Context) (result []model.Pong, err error)
	Page(ctx context.Context, in model.ServicePageIn) (out model.ServicePageOut, err error)
	Block(ctx context.Context, in model.ServiceBlockIn) (out model.ServiceBlockOut, err error)
}

func New(
	logger log.Log,
	cfg config.Config,
	metrics metric.ServiceMetric,
	utils utils.Utils,
) Service {
	return &service{
		logger,
		cfg,
		metrics,
		utils,
	}
}
