package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/gommon/color"

	"log"
)

// просмотр кофигурационных файлов
func (c *Lib) Ls() (result []map[string]string) {

	fmt.Println("List configuration:")
	fmt.Printf("%-29s%-17s%-17s%-17s%-16s%-30s%-60s\n", color.Green("DOMAIN"), color.Green("API"), color.Green("GUI"), color.Green("PROXY"), color.Green("APP"), color.Green("CACHE"), color.Green("CONFIG ID"))
	sep := string(filepath.Separator)

	// может работать много прокси, поэтому обходим конфигурационные файлы и ищем рабочие прокси
	pathFolder := c.RootDir() + sep + "upload" //+ sep + c.State["domain"] + sep + "ini"
	folders, err := ioutil.ReadDir(pathFolder)
	if err != nil {
		c.Logger.Panic(err)
		return
	}

	// пробегаем текущую папку и считаем совпадание признаков
	for _, obj := range folders {
		if obj.IsDir() {

			nextPath := pathFolder + sep + obj.Name() + sep + "ini"
			// читаю вложенную директорию с конфигурациями
			files, err := ioutil.ReadDir(nextPath)

			if err != nil {
				continue
			}

			for _, file := range files {
				conf, _, err := ReadConf(file.Name())
				result = append(result, conf)

				if err == nil {
					port_api := conf["port_api"]
					if port_api == "" {
						port_api = "auto"
					}
					port_gui := conf["port_gui"]
					if port_gui == "" {
						port_gui = "auto"
					}
					port_proxy := conf["port_proxy"]
					if port_proxy == "" {
						port_proxy = "-"
					} else {
						if conf["hosts"] == "" {
							port_proxy = "disable"
						}
					}

					port_app := conf["port_app"]
					if port_app == "" {
						port_app = "-"
					} else {
						port_api = "-"
						port_gui = "-"
						port_proxy = "-"
					}

					fileConfig := file.Name()[:len(file.Name())-5]
					fmt.Printf("%-20.20s%-8s%-8s%-8s%-8s%-20s%-40.40s\n", conf["domain"], port_api, port_gui, port_proxy, port_app, conf["cache_pointvalue"], fileConfig)
				}
			}

		}
	}

	fmt.Println()

	return result
}

// просмотр запущенных сервисов
// format - тип вывода
// terminal - пишем в терминал списко процессов,
// pid - список пидов,
// full - полный слайс значений как для терминала, но в структуре
// raw - слайс всех полученных PidRegistry ответов
func (c *Lib) Ps(format string) (pids []string, services map[string][][]string, raw []map[string]map[string][]string, err error) {
	var PidRegistry = map[string]map[string][]string{}
	var finish = map[string][]string{}
	sep := string(filepath.Separator)

	if format == "terminal" {
		fmt.Printf("%-29s%-17s%-17s%-60s\n", color.Green("DOMAIN"), color.Green("PID"), color.Green("PORT"), color.Green("CONFIG ID"))
	}

	// может работать много прокси, поэтому обходим конфигурационные файлы и ищем рабочие прокси
	pathFolder := c.RootDir() + sep + "upload" //+ sep + c.State["domain"] + sep + "ini"
	folders, err := ioutil.ReadDir(pathFolder)
	if err != nil {
		log.Panic(err)
		return
	}

	// пробегаем текущую папку и считаем совпадание признаков
	for _, obj := range folders {
		if obj.IsDir() {

			nextPath := pathFolder + sep + obj.Name() + sep + "ini"
			// читаю вложенную директорию с конфигурациями
			files, err := ioutil.ReadDir(nextPath)
			if err == nil {

				for _, file := range files {
					conf, _, err := ReadConf(file.Name())

					if err == nil {
						// смотрим настройки на наличик возможно досутпного прокси
						if conf["hosts"] != "" && conf["port_proxy"] != "" {

							// получаем список доступных на данном прокси запущенных приложений
							// ПЕРЕДЕЛАТЬ!!! слишком много реализаций Curl - сделать ревью!!!! убрать дубли и вынести в lib
							_, err = c.Curl("GET", "http://localhost:"+conf["port_proxy"]+"/pid", "", &PidRegistry, map[string]string{})

							// просто слайс всех PidRegistry
							raw = append(raw, PidRegistry)

							if err == nil && len(PidRegistry) != 0 {

								if format == "terminal" {
									fmt.Printf("%-80s\n", color.Yellow("Running proxy: ")+color.Yellow(conf["port_proxy"])+" - "+file.Name())
								}

								domain := ""
								for k, v := range PidRegistry {
									domain = k
									for k1, v1 := range v {
										domain = k + "/" + k1

										// добавляем в структуру полученне значения
										// бывает указаны в несколькхи конфигах порты прокси, чтобы не дублировались пишем сначала в структуру
										if _, found := finish[domain]; !found {
											finish[domain] = v1
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// выводим структуру значений запущенных процессов
	var k3 = []string{}
	var wpids = []string{}
	var slice = map[string][][]string{}

	for kf, vf := range finish {
		for _, v4 := range vf {
			k3 = strings.Split(v4, ":")

			if len(k3) > 0 {
				if format == "terminal" {
					fmt.Printf("%-20.20s%-8s%-8s%-41.41s\n", kf, k3[0], k3[2], k3[1])
				}
				if format == "pid" {
					wpids = append(wpids, k3[0])
				}
			}

			if format == "full" {
				slice[kf] = append(slice[kf], k3)
			}
		}

	}

	if format == "terminal" {
		fmt.Println()
	}

	return wpids, slice, raw, err
}

// завершение процесса
func (c *Lib) Stop(pid int) error {
	var sig os.Signal
	sig = os.Kill
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = p.Signal(sig)
	return err
}

// завершение всех процессов для текущей конфигурации
// config - ид-конфигурации
func (c *Lib) PidsByConfig(config string) (result []string, err error) {

	_, fullresult, _, _ := c.Ps("full")

	// получаем pid для переданной конфигурации
	for _, v1 := range fullresult {
		for _, v := range v1 {
			configfile := v[1] // файл
			idProcess := v[0]  // pid

			if config == configfile {
				result = append(result, idProcess)
			}

			if err != nil {
				c.Logger.Error(err, "Error stopped process config:", config)
				fmt.Println("Error stopped process config:", config, ", err:", err)
			}
		}
	}

	return
}

// получаем строки пидов подходящих под условия, в котором:
// domain - название проекта (домен)
// alias - название алиас-сервиса (gui/api/proxy и тд - то, что в мап-прокси идет второй частью адреса)
// если алиас явно не задан, то он может быть получен из домена
func (c *Lib) PidsByAlias(domain, alias string) (result []string, err error) {

	if domain == "" {
		domain = "all"
	}
	if alias == "" {
		alias = "all"
	}

	// можем в домене передать полный путь с учетом алиаса типа buildbox/gui
	// в этом случае алиас если он явно не задан заполним значением алиаса полученного из домена
	splitDomain := strings.Split(domain, "/")
	if len(splitDomain) == 2 {
		domain = splitDomain[0]
		alias = splitDomain[1]
	}
	_, _, raw, _ := c.Ps("full")

	// получаем pid для переданной конфигурации
	for _, pidRegistry := range raw {
		for d, v1 := range pidRegistry {
			// пропускаем если точное сравнение и не подоходит
			if domain != "all" && d != domain {
				continue
			}

			for a, v2 := range v1 {
				// пропускаем если точное сравнение и не подоходит
				if alias != "all" && a != alias {
					continue
				}

				for _, v3 := range v2 {
					k3 := strings.Split(v3, ":")
					idProcess := k3[0]  // pid
					result = append(result, v3)

					if err != nil {
						c.Logger.Error(err, "Error stopped process: pid:", idProcess)
						fmt.Println("Error stopped process: pid:", idProcess, ", err:", err)
					}
				}
			}
		}
	}

	return
}

// перезагрузка процесса по Pid/домену
//func (c *Lib) Reload(pid string) (err error) {
//	sep := string(filepath.Separator)
//	pidDone := ""
//
//	_, fullresult, _, _ := Ps("full")
//
//	for domain, v1 := range fullresult {
//		for _, v := range v1 {
//			configfile := v[1] // файл
//			idProcess := v[0]  // pid
//
//			// pid не передан, перезагружаем все процессы
//			if pid == "all" {
//
//				// если данный пид ранее не обрабатывался, то рестартуем процесс
//				if !strings.Contains(pidDone, idProcess) {
//					pidI, err := strconv.Atoi(idProcess)
//					err = Stop(pidI)
//					if err == nil {
//						RunProcess(CurrentDir()+sep+"buildbox", configfile, "start", "service")
//					}
//				}
//				// сохраняем обработанный пид (чтобы повторно не релоадить)
//				pidDone = pidDone + "," + idProcess
//
//			} else {
//				// перезагружаем только переданный pid
//				if !strings.Contains(pidDone, idProcess) {
//					if pid == idProcess || pid == domain {
//						pidI, err := strconv.Atoi(idProcess)
//						err = Stop(pidI)
//						if err == nil {
//							RunProcess(CurrentDir()+sep+"buildbox", configfile, "start", "service")
//						}
//					}
//				}
//				// сохраняем обработанный пид (чтобы повторно не релоадить)
//				pidDone = pidDone + "," + pid
//
//			}
//		}
//
//	}
//
//	return err
//}

// уничтожить все процессы
func (c *Lib) Destroy() (err error) {
	pids, _, _, _ := c.Ps("pid")
	for _, v := range pids {
		pi, err := strconv.Atoi(v)
		if err == nil {
			Stop(pi)
		}
	}
	return err
}

// инициализация приложения
func (c *Lib) Install() (err error) {

	// 1. задание переменных окружения
	os.Setenv("BBPATH", CurrentDir())

	//var rootPath = os.Getenv("BBPATH")

	//fmt.Println(rootPath)
	//path, _ := os.LookupEnv("BBPATH")
	//fmt.Print("BBPATH: ", path)

	// 2. копирование файла запуска в /etc/bin
	//src := "./buildbox"
	//dst := "/usr/bin/buildbox"
	//
	//in, err := os.Open(src)
	//if err != nil {
	//	return err
	//}
	//defer in.Close()
	//
	//out, err := os.Create(dst)
	//if err != nil {
	//	return err
	//}
	//defer out.Close()
	//
	//_, err = io.Copy(out, in)
	//if err != nil {
	//	return err
	//}
	//return out.Close()

	return err
}
