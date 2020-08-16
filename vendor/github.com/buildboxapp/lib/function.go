package lib

import (
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"os/exec"
	"strconv"
	"syscall"

	"crypto/sha1"
	"encoding/hex"
	"github.com/labstack/gommon/color"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//func (c *Lib) Init(output io.Writer, urlgui, urlapi string) {
//	c.Logger.Output = output
//	c.UrlGUI = urlgui
//	c.UrlAPI = urlapi
//}

func (c *Lib) ResponseJSON(w http.ResponseWriter, objResponse interface{}, status string, error error, metrics interface{}) {

	if w == nil {
		return
	}

	errMessage := RestStatus{}
	st, found := StatusCode[status]
	if found {
		errMessage = st
	} else {
		errMessage = StatusCode["NotStatus"]
	}

	objResp := &Response{}
	if error != nil {
		errMessage.Error = fmt.Sprint(error)
	}

	// Metrics
	b1, _ := json.Marshal(metrics)
	var metricsR Metrics
	json.Unmarshal(b1, &metricsR)
	if metrics != nil {
		objResp.Metrics = metricsR
	}

	objResp.Status = errMessage
	objResp.Data = objResponse

	// формируем ответ
	out, err := json.Marshal(objResp)
	if err != nil {
		log.Printf("%s", err)
	}

	//WriteFile("./dump.json", out)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(out)
}

// стартуем сервис из конфига
func (c *Lib) RunProcess(fileConfig, workdir, file, command, message string) (err error) {
	var out []byte

	if fileConfig == "" {
		fmt.Println(color.Red("ERROR!") + " Configuration file is not found.\n")
		return
	}

	if command == "" {
		command = "start"
	}

	done := color.Green("OK")
	fail := color.Red("FAIL")
	fileStart := workdir + "/" + file

	cmd := exec.Command(fileStart, command, "--config", fileConfig)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	//stdout, err := cmd.StdoutPipe()
	//if err != nil {
	//	log.Fatal(err)
	//}
	if err := cmd.Start(); err != nil {
		fmt.Println(err, ". Error run ", message, ": cmd - ", cmd)
		c.Logger.Error(err, ". Error ", message, ": cmd - ", cmd)
		return err
	}
	//var person interface{}
	//if err := json.NewDecoder(stdout).Decode(&person); err != nil {
	//	log.Fatal(err)
	//}
	//if err := cmd.Wait(); err != nil {
	//	log.Fatal(err)
	//}

	//fmt.Printf("%s is %d years old\n", person)

	if err != nil {
		fmt.Printf("%s Starting %s: %s\n%s", fail, message, err, string(out))
		c.Logger.Error(err, "from starting: ", message, "-", file, "(", command, ")", string(out))
		return err
	}

	fmt.Printf("%s Starting %s (pid: %s) \n", done, message, strconv.Itoa(cmd.Process.Pid))
	c.Logger.Info("Starting: ", message, "-", file, "(", command, ")", string(out))

	return err
}

// останавливаем сервис по порту
//func (c *Lib) StopProcess(workdir, fileConfig, message string) {
//
//	if fileConfig == "" {
//		fmt.Println(color.Red("ERROR!") + " Configuration file is not found.\n")
//		return
//	}
//
//	var err error
//	done := color.Yellow("OK")
//	fail := color.Red("FAIL")
//	fileStart := workdir + "/buildbox"
//
//	cmd := exec.Command(fileStart, "stop", "--config", fileConfig)
//	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
//	err = cmd.Start()
//	if err != nil {
//		fmt.Printf("%s Exist %s: %s\n", message, fail, err)
//		return
//	}
//
//	fmt.Printf("%s Exist %s: %s\n", done, message, cmd.Process.Pid)
//	return
//}

//////////////////////////////////////////////////////////////////
////////////////////////// СЕРВИСНЫЕ ФУНКЦИИ /////////////////////
//////////////////////////////////////////////////////////////////

// читаем файл конфигурации и возвращаем
// объект конфига, джейсон-конфига и ошибку
func (c *Lib) ReadConf(configfile string) (conf map[string]string, confjson string, err error) {
	fpath := ""
	fileName := configfile

	// если нет / значит запускаем внутри корневой папки и формируем путь автоматически
	// иначе передается полный путь к файлу, который используем без изменений
	if !strings.Contains(configfile, "/") {
		fpath += "ini/"
		fpath += configfile
		if !strings.Contains(configfile, ".json") {
			fpath += ".json"
		}
		fileName = CurrentDir() + "/" + fpath
	}

	confJson, err := c.ReadFile(fileName)

	if err != nil {
		return nil, "", err
	}
	err = json.Unmarshal([]byte(confJson), &conf)

	return conf, confJson, err
}

// получаем конфигурацию по-умолчанию для сервера (перебираем конфиги и ищем первый у которого default=on)
func (c *Lib) DefaultConfig() (fileConfig string, err error) {
	fpath := ""
	if !strings.Contains(fileConfig, "/") {
		fpath = CurrentDir() + "/ini"
	}

	c.Logger.Info("Search DefaultConfig from : ", fpath)

	files, err := ioutil.ReadDir(fpath)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		conf, _, err := ReadConf(file.Name())
		if err == nil {
			d := conf["default"]
			if d != "" {
				fileConfig = file.Name()
				continue
			}

		}
	}

	c.Logger.Info("Search DefaultConfig result : ", fileConfig)

	return fileConfig, err
}

// определяем текущий каталог для первого запуска, чтобы прочитать файл с конфигурацией
func (c *Lib) CurrentDir() string {
	// путь к шаблонам при запуске через командную строку
	var runDir, _ = os.Getwd()
	var currentDir = filepath.Dir(os.Args[0]) // если запускать с goland отдает темповую папку (заменяем)
	if currentDir != runDir {
		currentDir = runDir
	}
	return currentDir
}

// корневую директорию (проверяем признаки в текущей директории + шагом вверх + шагом вниз)
func (c *Lib) RootDir() string {
	// путь к шаблонам при запуске через командную строку
	var runDir, _ = os.Getwd()
	var currentDir = filepath.Dir(os.Args[0]) // если запускать с goland отдает темповую папку (заменяем)
	if currentDir != runDir {
		currentDir = runDir
	}

	// признаки рутовой директории - наличие файла buildbox (стартового (не меняется)
	// наличие директорий ini + bin +

	return currentDir
}

func (c *Lib) Hash(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	return sha1_hash
}

func (c *Lib) PanicOnErr(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		panic(err)
	}
}

func (c *Lib) UUID() string {
	stUUID := uuid.NewV4()
	return stUUID.String()
}
