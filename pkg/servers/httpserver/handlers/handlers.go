package handlers

import (
	"github.com/buildboxapp/app/pkg/config"
	"github.com/buildboxapp/app/pkg/service"
	bblog "github.com/buildboxapp/lib/log"
	"net/http"
)

type handlers struct {
	service service.Service
	logger bblog.Log
	cfg config.Config
}

type Handlers interface {
	Alive(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	Page(w http.ResponseWriter, r *http.Request)
	Block(w http.ResponseWriter, r *http.Request)
}

func New(
	service service.Service,
	logger bblog.Log,
	cfg config.Config,
) Handlers {
	return &handlers{
		service,
		logger,
		cfg,
	}
}