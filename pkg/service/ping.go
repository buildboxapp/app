package service

import (
	"context"
	"encoding/json"
	"github.com/buildboxapp/app/pkg/model"
	"os"
	"strconv"
	"strings"
)


// Ping ...
func (s *service) Ping(ctx context.Context) (result []model.Pong, err error) {
	pp := strings.Split(cfg.Domain, "/")
	name := "ru"
	version := "ru"

	if len(pp) == 1 {
		name = pp[0]
	}
	if len(pp) == 2 {
		name = pp[0]
		version = pp[1]
	}

	pg, _ := strconv.Atoi(cfg.PortGui)
	pid := strconv.Itoa(os.Getpid())+":"+cfg.UidGui
	state, _ := json.Marshal(ServiceMetrics.Get())

	var r = []model.Pong{
		{name, version, "run",pg, pid, string(state),s.cfg.ReplicasGui.Value},
		{"assets", version, "run",pg, pid, string(state), s.cfg.ReplicasGui.Value},
	}

	return r, err
}