package api

import (
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/app/pkg/utils"
	"github.com/buildboxapp/lib/log"
	"github.com/buildboxapp/lib/metric"
	"strings"
	"time"
)

type api struct {
	logger  log.Log
	utils   utils.Utils
	cfg 	model.Config
	metric 	metric.ServiceMetric
}

type Api interface {
	Obj
}

func New(logger log.Log, utils utils.Utils, cfg model.Config, metric metric.ServiceMetric) Api {
	return &api{
		logger,
		utils,
		cfg,
		metric,
	}
}

type Obj interface {
	ObjGet(uids string) (result model.ResponseData, err error)
	ObjCreate(bodymap map[string]string) (result model.ResponseData, err error)
	ObjAttrUpdate(uid, name, value, src, editor string) (result model.ResponseData, err error)
	LinkGet(tpl, obj, mode, short string) (result model.ResponseData, err error)
	Query(query, method, bodyJSON string, response interface{}) (result interface{}, err error)
}

// результат выводим в объект как при вызове Curl
func (o *api) Query(query, method, bodyJSON string, response interface{}) (result interface{}, err error) {
	urlc := o.cfg.UrlApi + "/query/"+query
	urlc = strings.Replace(urlc, "//query", "/query", 1)

	result, err = o.utils.Curl(method, urlc, bodyJSON, &response, map[string]string{})
	return result, err
}

func (o *api) ObjGet(uids string) (result model.ResponseData, err error) {
	urlc := o.cfg.UrlApi + "/query/obj?obj="+uids
	urlc = strings.Replace(urlc, "//query", "/query", 1)

	_, err = o.utils.Curl("GET", urlc, "", &result, map[string]string{})
	return result, err
}

func (o *api) LinkGet(tpl, obj, mode, short string) (result model.ResponseData, err error) {
	urlc := o.cfg.UrlApi + "/link/get?source="+tpl+"&mode="+mode+"&obj="+obj+"&short="+short
	urlc = strings.Replace(urlc, "//link", "/link", 1)

	_, err = o.utils.Curl("GET", urlc, "", &result, map[string]string{})

	return result, err
}

// изменение значения аттрибута объекта
func (a *api) ObjAttrUpdate(uid, name, value, src, editor string) (result model.ResponseData, err error)  {

	post := map[string]string{}
	thisTime := fmt.Sprintf("%v", time.Now().UTC())

	post["uid"] = uid
	post["element"] = name
	post["value"] = value
	post["src"] = src
	post["rev"] = thisTime
	post["path"] = ""
	post["token"] = ""
	post["editor"] = editor

	dataJ, _ := json.Marshal(post)
	result, err = a.Element("update", string(dataJ))

	return result, err
}

// ПЕРЕДЕЛАТЬ на понятные пути в ORM
// сделано так для совместимости со старой версией GUI
func (a *api) Element(action, body string) (result model.ResponseData, err error) {
	_, err = a.utils.Curl("POST", a.cfg.UrlApi + "/element/"+action, body, &result, map[string]string{})

	return result, err
}

func (a *api) ObjCreate(bodymap map[string]string) (result model.ResponseData, err error) {
	body, _ := json.Marshal(bodymap)
	_, err = a.utils.Curl("POST", a.cfg.UrlApi + "/objs?format=json", string(body), &result, map[string]string{})

	return result, err
}
