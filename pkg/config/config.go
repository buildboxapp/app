// преобразование типов для чтения конфигурации из файла
package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/buildboxapp/lib"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/gommon/color"
	"os"
	"strings"
)

var warning = color.Red("[Fail]")

// читаем конфигурации
func (c *Config) load(configfile string) (err error) {
	fileName := ""
	cfgfile := ""

	rootDir, err := lib.RootDir()
	startDir := rootDir + sep + "upload"
	if err := envconfig.Process("", c); err != nil {
		fmt.Printf("%s Error load enviroment: %s (configfile: %s)\n", warning, err, cfgfile)
		os.Exit(1)
	}

	// временно, пока не перешли полностью на cfg (позже удалить)
	if len(configfile) > 5 {
		if configfile[len(configfile)-5:] != ".json" {
			configfile = configfile + ".json"
		}
	}

	if fileName, err = lib.ReadConfAction(startDir, configfile, false); err != nil {
		fmt.Printf("%s Error load enviroment: %s (configfile: %s)\n", warning, err, configfile)
		os.Exit(1)
	}

	if len(fileName) > 5 {
		if fileName[len(fileName)-5:] == ".json" {
			cfgfile = fileName[:len(fileName)-5]
		}
		cfgfile = cfgfile + ".cfg"
	}
	if _, err = toml.DecodeFile(cfgfile, &c); err != nil {
		fmt.Printf("%s Error: %s (configfile: %s)\n", warning, err, cfgfile)
		os.Exit(1)
	}

	return err
}

// формируем ClientPath из Domain
func (c *Config) SetClientPath()  {
	pp := strings.Split(c.Domain, "/")
	name := "buildbox"
	version := "gui"

	if len(pp) == 1 {
		name = pp[0]
	}
	if len(pp) == 2 {
		name = pp[0]
		version = pp[1]
	}
	c.ClientPath = "/" + name + "/" + version

	return
}

func New(configfile string) Config {
	var cfg = Config{}
	return cfg
}