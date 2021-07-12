package session

import (
	"github.com/buildboxapp/app/pkg/api"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/lib/log"
	"sync"
)

type session struct {
	logger  log.Log
	cfg 	model.Config
	api 	api.Api

	Registry SessionRegistry
}

type SessionRegistry struct {
	Mx sync.Mutex
	M map[string]SessionRec
}

type SessionRec struct {
	DeadTime int64
	Profile model.ProfileData
}

type Session interface {
	Get(sessionID string) (profile model.ProfileData, err error)
	Delete(sessionID string) (err error)
	Set(token *model.Token) (err error)
	Profile(token *model.Token) (profile model.ProfileData, err error)
	List() (result map[string]SessionRec)
}

func New(logger log.Log, cfg model.Config, api api.Api) Session {
	registrySession := SessionRegistry{}

	return &session{
		logger,
		cfg,
		api,
		registrySession,
	}
}