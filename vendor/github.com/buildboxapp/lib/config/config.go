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

const sep = string(os.PathSeparator)
var warning = color.Red("[Fail]")

// читаем конфигурации
// получаем только название конфигурации
// 1. поднимаемся до корневой директории
// 2. от нее ищем полный путь до конфига
// 3. читаем по этому пути
func Load(configname string, pointToCfg interface{}) (err error) {
	if err := envconfig.Process("", pointToCfg); err != nil {
		fmt.Printf("%s Error load default enviroment: %s\n", warning, err)
		err = fmt.Errorf("Error load default enviroment: %s", err)
		return err
	}

	// 1.
	rootDir, err := lib.RootDir()
	if err != nil {
		return err
	}

	// 2.
	confidPath, err := fullPathConfig(rootDir, configname)
	if err != nil {
		return err
	}

	// 3.
	err = read(confidPath, pointToCfg)
	if err != nil {
		return err
	}

	return err
}

// получаем путь от переданной директории
// если defConfig = true - значит ищем конфигурацию по-умолчанию
func fullPathConfig(rootDir, configuration string) (configPath string, err error) {
	var nextPath string
	directory, err := os.Open(rootDir)
	if err != nil {
		return "", err
	}
	defer directory.Close()

	objects, err := directory.Readdir(-1)
	if err != nil {
		return "", err
	}

	// пробегаем текущую папку и считаем совпадание признаков
	for _, obj := range objects {
		nextPath = rootDir + sep + obj.Name()
		if obj.IsDir() {
			dirName := obj.Name()

			// не входим в скрытые папки
			if dirName[:1] != "." {
				configPath, err = fullPathConfig(nextPath, configuration)
				if configPath != "" {
					return configPath, err // поднимает результат наверх
				}
			}

		} else {
			if configuration == "default" { // проверяем на получение конфигурации по-умолчанию
				if strings.Contains(nextPath, ".cfg") {
					//confJson, err := ReadFile(nextPath)
					//err = json.Unmarshal([]byte(confJson), &conf)
					//if err == nil {
					//	d := conf["default"]
					//	if d == "checked" {
					//		return nextPath, err
					//	}
					//}
				}
			} else {
				if !strings.Contains(nextPath, "/.") {
					// проверяем только файлы конфигурации (игнорируем .json)
					if strings.Contains(obj.Name(), configuration + ".cfg") {
						return nextPath, err
					}
				}
			}
		}
	}

	return configPath, err
}

// Читаем конфигурация по заданному полному пути
func read(configfile string, cfg interface{}) (err error) {
	configfileSplit := strings.Split(configfile, ".")
	if len(configfile) == 0 {
		return fmt.Errorf("%s", "Error. Configfile is empty.")
	}
	if len(configfileSplit) == 1 {
		configfile = configfile + ".cfg"
	}
	if _, err = toml.DecodeFile(configfile, cfg); err != nil {
		fmt.Printf("%s Error: %s (configfile: %s)\n", warning, err, configfile)
	}

	return err
}