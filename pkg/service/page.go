package service

import (
	"context"
	"encoding/json"
	"github.com/buildboxapp/app/pkg/model"
	"os"
	"strconv"
	"strings"
)


// Page ...
func (s *service) Page(ctx context.Context) (result []model.Pong, err error) {
	pp := strings.Split(s.cfg.Domain, "/")
	name := "ru"
	version := "ru"

	if len(pp) == 1 {
		name = pp[0]
	}
	if len(pp) == 2 {
		name = pp[0]
		version = pp[1]
	}

	pg, _ := strconv.Atoi(s.cfg.PortApp)
	pid := strconv.Itoa(os.Getpid())+":"+s.cfg.UidApp
	state, _ := json.Marshal(s.metrics.Get())

	var r = []model.Pong{
		{name, version, "run",pg, pid, string(state),s.cfg.ReplicasApp.Value},
	}

	return r, err
}