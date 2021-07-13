package session

import (
	"github.com/buildboxapp/app/pkg/api"
	"github.com/buildboxapp/app/pkg/iam"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/lib/log"
	"sync"
)

type session struct {
	logger  log.Log
	cfg 	model.Config
	api 	api.Api
	iam 	iam.IAM

	Registry SessionRegistry
}

type SessionRegistry struct {
	Mx sync.Mutex
	M map[string]SessionRec
}

type SessionRec struct {
	UID string	`json:"uid"`
	DeadTime int64	`json:"dead_time"`
	Profile model.ProfileData `json:"profile"`
}

type Session interface {
	Get(sessionID string) (profile model.ProfileData, err error)
	Delete(sessionID string) (err error)
	Set(token *model.Token) (err error)
	List() (result map[string]SessionRec)
}

func New(logger log.Log, cfg model.Config, api api.Api, iam iam.IAM) Session {
	registrySession := SessionRegistry{}

	return &session{
		logger,
		cfg,
		api,
		iam,
		registrySession,
	}
}