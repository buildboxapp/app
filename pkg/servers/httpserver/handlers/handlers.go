package handlers

import (
	"encoding/json"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/app/pkg/service"
	bblog "github.com/buildboxapp/lib/log"
	"net/http"
)

type handlers struct {
	service service.Service
	logger  bblog.Log
	cfg     model.Config
}

type Handlers interface {
	Alive(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	Page(w http.ResponseWriter, r *http.Request)
	Block(w http.ResponseWriter, r *http.Request)
	Cache(w http.ResponseWriter, r *http.Request)
}

func (h *handlers) transportResponse(w http.ResponseWriter, response interface{}) (err error)  {
	w.WriteHeader(200)
	d, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(403)
	}
	w.Write(d)
	return err
}

func (h *handlers) transportError(w http.ResponseWriter, code int, error error, message string) (err error)  {
	var res = model.Response{}

	res.Status.Error = error
	res.Status.Description = message
	d, err := json.Marshal(res)

	h.logger.Error(err, message)

	w.WriteHeader(code)
	w.Write(d)
	return err
}

func (h *handlers) transportResponseHTTP(w http.ResponseWriter, response string) (err error)  {
	w.WriteHeader(200)

	if err != nil {
		w.WriteHeader(403)
	}
	w.Write([]byte(response))
	return err
}


func New(
	service service.Service,
	logger bblog.Log,
	cfg model.Config,
) Handlers {
	return &handlers{
		service,
		logger,
		cfg,
	}
}