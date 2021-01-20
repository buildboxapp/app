package service

import (
	"context"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/lib/log"
)

type service struct {
	state map[string]string
	logger log.Log
}

// Service interface
type Service interface {
	Ping(ctx context.Context) (result model.Pong, err error)
	ServiceAction(action, config, domain, alias string) (result string)
}

func New(
	state map[string]string,
	logger log.Log,
) Service {
	return &service{
		state,
		logger,
	}
}
