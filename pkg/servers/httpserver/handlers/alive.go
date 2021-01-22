package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Alive godoc
// @Summary alive
// @Description check application health
// @Produce  plain
// @Success 200 {string} string	"OK"
// @Router /alive [get]
func (h *handlers) Alive(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	curVersion := fmt.Sprintf("<p>HTTP OK. v%s</p>", h.cfg.AppVersionPointvalue)
	curConfig, _ := json.Marshal(h.cfg)

	result := fmt.Sprintf("<html><body>Version: %s/n/nConfig: /n%s</body></html>", curVersion, curConfig)
	_, _ = w.Write([]byte(result))
}