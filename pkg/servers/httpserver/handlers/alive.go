package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gomarkdown/markdown"

	"github.com/buildboxapp/gui/pkg/config"
)

// Alive godoc
// @Summary alive
// @Description check application health
// @Produce  plain
// @Success 200 {string} string	"OK"
// @Router /alive [get]
func Alive(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	curVersion := fmt.Sprintf("<p>HTTP OK. v%s</p>", getVersion())
	changelogHTML := ""
	var env = os.Getenv("ENVIRONMENT")
	if env != config.ProdEnvironment {
		changelogHTML = renderChangelog()
	} else {
		changelogHTML = fmt.Sprintf("<p>Env: %s</p>", env)
	}

	result := fmt.Sprintf("<html><body>%s%s</body></html>", curVersion, changelogHTML)
	_, _ = w.Write([]byte(result))
}

func renderChangelog() string {
	data, _ := ioutil.ReadFile("CHANGELOG.md")
	output := markdown.ToHTML(data, nil, nil)
	return string(output)
}

var serviceVersion = "dev"

func getVersion() string {
	return serviceVersion
}
