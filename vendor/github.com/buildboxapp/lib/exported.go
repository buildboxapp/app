package lib

import "net/http"

var lb *Lib

// HTTP-functions
func Curl(method, urlc, bodyJSON string, response interface{}) (result interface{}, err error) {
	return lb.Curl(method, urlc, bodyJSON, response)
}


// CLI-functions
func Ls() {
	lb.Ls()
}

func Ps(format string) (pids []string, services map[string][][]string, raw []map[string]map[string][]string, err error)  {
	return lb.Ps(format)
}

func Stop(pid int) error {
	return lb.Stop(pid)
}

func StopByConfig(config string) error {
	return lb.StopByConfig(config)
}

func Reload(pid string) error {
	return lb.Reload(pid)
}

func Destroy() error {
	return lb.Destroy()
}

func Install() error {
	return lb.Install()
}


// FUNCTION-functions
func ResponseJSON(w http.ResponseWriter, objResponse interface{}, status string, error error, metrics interface{}) {
	lb.ResponseJSON(w, objResponse, status, error, metrics)
}

func RunProcess(fileConfig, workdir, file, command, message string) error {
	return lb.RunProcess(fileConfig, workdir, file, command, message)
}

func ReadConf(configfile string) (conf map[string]string, confjson string, err error) {
	return lb.ReadConf(configfile)
}

func DefaultConfig() (fileConfig string, err error) {
	return lb.DefaultConfig()
}

func CurrentDir() string {
	return lb.CurrentDir()
}


// FILES-functions
func CreateFile(path string) {
	lb.CreateFile(path)
}

func isError(err error) bool {
	return lb.isError(err)
}

func WriteFile(path string, data []byte) (err error) {
	return lb.WriteFile(path, data)
}

func ReadFile(path string) (result string, err error) {
	return lb.ReadFile(path)
}


// Прочие фукнции
func Hash(str string) string {
	return lb.Hash(str)
}

func PanicOnErr(err error) {
	lb.PanicOnErr(err)
}

func UUID() string {
	return lb.UUID()
}



// zipit("/tmp/documents", "/tmp/backup.zip")
func Zip(source, target string) error {
	return lb.Zip(source, target)
}

// unzip("/tmp/report-2015.zip", "/tmp/reports/")
func Unzip(archive, target string) error {
	return lb.Unzip(archive, target)
}


// Функции шифрования/расшифрования
func Encrypt(key []byte, text string) (string, error) {
	return lb.Encrypt(key, text)
}

func Decrypt(key []byte, text string) (string, error) {
	return lb.Decrypt(key, text)
}

