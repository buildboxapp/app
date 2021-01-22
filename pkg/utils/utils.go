package utils

import (
	"github.com/buildboxapp/app/pkg/config"
	"github.com/buildboxapp/lib/log"
)

type utils struct {
	cfg config.Config
	logger log.Log
}

type Utils interface {
	AddressProxy()
	Curl(method, urlc, bodyJSON string, response interface{}) (result interface{}, err error)
}


func New(cfg config.Config, logger log.Log) Utils {
	return &utils{
		cfg,
		logger,
	}
}