package service

import (
	"context"
	"github.com/buildboxapp/app/pkg/model"
)

// Alive ...
func (s *service) Alive(ctx context.Context) (out model.AliveOut, err error) {
	out.Config = s.cfg
	out.Cache = s.cache.Active()

	return
}
