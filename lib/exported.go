package app_lib

import (
	"net/http"
	"sync"
	"html/template"
)

var lb *App

// DOGFUNC - функции
func SplitIndex(arg []string) (result string) {
	return lb.SplitIndex(arg)
}

func Time(arg []string) (result string) {
	return lb.Time(arg)
}

func FuncURL(r *http.Request, arg []string) (result string) {
	return lb.FuncURL(r, arg)
}

func TplValue(values map[string]interface{}, arg []string) (result string) {
	return lb.TplValue(values, arg)
}

func ConfigValue(arg []string) (result string) {
	return lb.ConfigValue(arg)
}

func Path(d []Data, arg []string) (result string) {
	return lb.Path(d, arg)
}

func UserObj(r *http.Request, arg []string) (result string) {
	return lb.UserObj(r, arg)
}

func UserProfile(r *http.Request, arg []string) (result string) {
	return lb.UserProfile(r, arg)
}

func UserRole(r *http.Request, arg []string) (result string) {
	return lb.UserRole(r, arg)
}

func Obj(d []Data, arg []string) (result string) {
	return lb.Obj(d, arg)
}

func FieldValue(d []Data, arg []string) (result string) {
	return lb.FieldValue(d, arg)
}

func FieldSrc(d []Data, arg []string) (result string) {
	return lb.FieldSrc(d, arg)
}

func FieldSplit(d []Data, arg []string) (result string) {
	return lb.FieldSplit(d, arg)
}

func DogParse(p string, r *http.Request, queryData *[]Data, values map[string]interface{}) (result string) {
	return lb.DogParse(p, r, queryData, values)
}



// FUNCTION - функции
func hash(str string) string {
	return lb.hash(str)
}

func CreateFile(path string) {
	lb.CreateFile(path)
}

func isError(err error) bool {
	return lb.isError(err)
}

func WriteFile(path string, data []byte) {
	lb.WriteFile(path, data)
}

func Curl(method, urlc, bodyJSON string, response interface{}) (result interface{}, err error) {
	return lb.Curl(method, urlc, bodyJSON, response)
}

func ModuleBuild(block Data, r *http.Request, page Data, values map[string]interface{}, enableCache bool) (result ModuleResult) {
	return lb.ModuleBuild(block, r, page, values, enableCache)
}

func ModuleBuildParallel(p Data, r *http.Request, page Data, values map[string]interface{}, enableCache bool, buildChan chan ModuleResult, wg *sync.WaitGroup) {
	lb.ModuleBuildParallel(p, r, page, values, enableCache, buildChan, wg)
}

func ErrorModuleBuild(stat map[string]interface{}, buildChan chan ModuleResult, timerRun interface{}, errT error) {
	lb.ErrorModuleBuild(stat, buildChan, timerRun, errT)
}

func QueryWorker(queryUID, dataname string, source[]map[string]string, r *http.Request) interface{} {
	return lb.QueryWorker(queryUID, dataname, source, r)
}

func ErrorPage(err interface{}, w http.ResponseWriter, r *http.Request) {
	lb.ErrorPage(err, w, r)
}

func ModuleError(err interface{}, r *http.Request) template.HTML {
	return lb.ModuleError(err, r)
}

func GUIQuery(tquery string, r *http.Request) Response {
	return lb.GUIQuery(tquery, r)
}


// HANDLER - функции
func ProxyPing(w http.ResponseWriter, r *http.Request) {
	lb.ProxyPing(w, r)
}

func PIndex(w http.ResponseWriter, r *http.Request) {
	lb.PIndex(w, r)
}

func TIndex(w http.ResponseWriter, r *http.Request, Config map[string]string) template.HTML {
	return lb.TIndex(w, r, Config)
}

func BPage(r *http.Request, blockSrc string, objPage ResponseData, values map[string]interface{}) string {
	return lb.BPage(r, blockSrc, objPage, values)
}

func GetBlock(w http.ResponseWriter, r *http.Request) {
	lb.GetBlock(w, r)
}

func TBlock(r *http.Request, block Data, Config map[string]string) template.HTML {
	return lb.TBlock(r, block, Config)
}


// CACHE - функции
func SetCahceKey(r *http.Request, p Data) (key, keyParam string) {
	return lb.SetCahceKey(r, p)
}

func СacheGet(key string, block Data, r *http.Request, page Data, values map[string]interface{}, url string) (string, bool) {
	return lb.СacheGet(key, block, r, page, values, url)
}

func CacheSet(key string, block Data, page Data, value, url string) bool {
	return lb.CacheSet(key, block, page, value, url)
}

func cacheUpdate(key string, block Data, r *http.Request, page Data, values map[string]interface{}, url string) {
	lb.cacheUpdate(key, block, r, page, values, url)
}

func refreshTime(options Data) int {
	return lb.refreshTime(options)
}


// LOGGER - фукнция
func LoggerHTTP(inner http.Handler, name string) http.Handler {
	return lb.LoggerHTTP(inner, name)
}
