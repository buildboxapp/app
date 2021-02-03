package block

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/app/pkg/config"
	"github.com/buildboxapp/app/pkg/function"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/app/pkg/utils"
	"github.com/buildboxapp/lib/log"
	uuid2 "github.com/google/uuid"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type block struct {
	cfg config.Config
	logger log.Log
	utils utils.Utils
	function function.Function
	tplfunc function.TplFunc
}

type Block interface {
	Generate(in model.ServiceIn, block model.Data, page model.Data, values map[string]interface{}) (result model.ModuleResult)
	ErrorModuleBuild(stat map[string]interface{}, buildChan chan model.ModuleResult, timerRun interface{}, errT error)
	QueryWorker(queryUID, dataname string, source[]map[string]string, token, queryRaw, metod string, postForm url.Values) interface{}
	ErrorPage(err interface{}, w http.ResponseWriter, r *http.Request)
	ModuleError(err interface{}) template.HTML
	GUIQuery(tquery, token, queryRaw, method string, postForm url.Values) model.Response
}

func (b *block) Generate(in model.ServiceIn, block model.Data, page model.Data, values map[string]interface{}) (result model.ModuleResult) 	{
	var c bytes.Buffer
	var err error
	var t *template.Template

	result.Id = block.Id

	// обработка всех странных ошибок
	// ВКЛЮЧИТЬ ПОЗЖЕ!!!!
	defer func() {
		if er := recover(); er != nil {
			//ft, err := template.ParseFiles("./upload/control/templates/errors/503.html")
			//if err != nil {
			//	l.Logger.Error(err)
			//}
			//t = template.Must(ft, nil)

			result.Result = template.HTML(fmt.Sprint(er))
			result.Err = fmt.Errorf("%s", er)
			return
		}
	}()

	// заменяем в State localhost на адрес домена (если это подпроцесс то все норм, но если это корневой сервис,
	// то у него url_proxy - localhost и узнать реньше адрес мы не можем, ибо еще домен не инициировался
	// а значит подменяем localhost при каждом обращении к модулю
	if strings.Contains(b.cfg.UrlProxy, "localhost") {
		b.cfg.UrlProxy = "//" + in.Host
	}

	bl := model.Block{}
	bl.Mx.Lock()
	defer bl.Mx.Unlock()

	t1 := time.Now()
	stat := map[string]interface{}{}
	stat["start"] = t1
	stat["status"] = "OK"
	stat["title"] = block.Title
	stat["id"] = block.Id

	bl.Value = map[string]interface{}{}

	dataSet		:= make(map[string]interface{})
	dataname 	:= "default" // значение по-умолчанию (будет заменено в потоках)

	tplName, _ 	:= block.Attr("module", "src")
	tquery, _ 	:= block.Attr("tquery", "src")

	// //////////////////////////////////////////////////////////////////////////////
	// в блоке есть настройки поля расширенного фильтра, который можно добавить в самом блоке
	// дополняем параметры request-a, доп. параметрами, которые переданы через блок
	extfilter, _ 	:= block.Attr("extfilter", "value") // дополнительный фильтр для блока
	dv := []model.Data{block}
	extfilter, err = b.function.Exec(extfilter, &dv, bl.Value, in)
	if err != nil {
		b.logger.Error(err, "[Generate] Error parsing Exec from block.")
		fmt.Println("[Generate] Error parsing Exec from block: ", err)
	}

	extfilter = strings.Replace(extfilter, "?", "", -1)

	// парсим переденную строку фильтра
	m, err := url.ParseQuery(extfilter)
	if err != nil {
		b.logger.Error(err, "[Generate] Error parsing extfilter from block.")
		fmt.Println("[Generate] Error parsing extfilter from block.: ", err)
	}

	// добавляем в URL переданное значение из настроек модуля
	//var q url.Values
	var blockQuery = in.Query // Get a copy of the query values.
	for k, v := range m {
		blockQuery.Add(k, strings.Join(v, ",")) // Add a new value to the set. Переводим обратно в строку из массива
	}
	// //////////////////////////////////////////////////////////////////////////////

	tconfiguration , _ := block.Attr("configuration", "value")
	tconfiguration = strings.Replace(tconfiguration, "  ", "", -1)

	uuid := uuid2.New()

	if values != nil && len(values) != 0 {
		for k, v := range values {
			if _, found := bl.Value[k]; !found {
				bl.Value[k] = v
			}
		}
	}

	bl.Value["Rand"] =  uuid[1:6]  // переопределяем отдельно для каждого модуля
	bl.Value["URL"] = in.Url
	bl.Value["Prefix"] = "/" + b.cfg.Domain + "/" +b.cfg.PathTemplates
	bl.Value["Domain"] = b.cfg.Domain
	bl.Value["CDN"] = b.cfg.UrlFs
	bl.Value["Path"] = b.cfg.ClientPath
	bl.Value["Title"] = b.cfg.Title
	bl.Value["Form"] = in.Form
	bl.Value["RequestURI"] = in.RequestURI
	bl.Value["Referer"] = in.Referer
	bl.Value["Profile"] = in.Profile


	//fmt.Println("tconfiguration: block", block.Id, tconfiguration, "\n")

	// обработк @-функции в конфигурации
	dv = []model.Data{block}
	dogParseConfiguration, err := b.function.Exec(tconfiguration, &dv, bl.Value, in)
	if err != nil {
		mes := "[Generate] Error DogParse configuration: ("+fmt.Sprint(err)+") " + tconfiguration
		result.Result = b.ModuleError(mes)
		result.Err = err
		b.logger.Error(err, mes)
		return
	}

	//fmt.Println("dogParseConfiguration: block", block.Id, dogParseConfiguration, "\n")

	// конфигурация без обработки @-функции
	var confRaw map[string]model.Element
	if tconfiguration != "" {
		err = json.Unmarshal([]byte(tconfiguration), &confRaw)
	}
	if err != nil {
		mes := "[Generate] Error Unmarshal configuration: ("+fmt.Sprint(err)+") " + tconfiguration
		result.Result = b.ModuleError("[Generate] Error Unmarshal configuration: ("+fmt.Sprint(err)+") " + tconfiguration)
		result.Err = err
		b.logger.Error(err, mes)
		return
	}

	// конфигурация с обработкой @-функции
	var conf map[string]model.Element
	if dogParseConfiguration != "" {
		err = json.Unmarshal([]byte(dogParseConfiguration), &conf)
	}
	if err != nil {
		mes := "[Generate] Error json-format configurations: ("+fmt.Sprint(err)+") " + dogParseConfiguration
		result.Result = b.ModuleError("[Generate] Error json-format configurations: ("+fmt.Sprint(err)+") " + dogParseConfiguration)
		result.Err = err
		b.logger.Error(err, mes)
		return
	}

	// сформировал структуру полученных описаний датасетов
	var source []map[string]string
	if d, found := conf["datasets"]; found {
		rm, _ := json.Marshal(d.Source)
		err := json.Unmarshal(rm, &source)

		if err != nil {
			stat["status"] = "error"
			stat["description"] = fmt.Sprint(err)

			result.Result = b.ModuleError(err)
			result.Err = err
			result.Stat = stat
			mes := "[Generate] Error generate datasets."
			b.logger.Error(err, mes)
			return result
		}
	}


	// ПЕРЕДЕЛАТЬ НА ПАРАЛЛЕЛЬНЫЕ ПОТОКИ
	if tquery != "" {
		slquery := strings.Split(tquery,",")

		var name, uid string
		for _, queryUID := range slquery {

			// подставляем название датасета из конфигурации
			for _, v1 := range source {
				if _, found := v1["name"]; found {
					name = v1["name"]
				}
				if _, found := v1["uid"]; found {
					uid = v1["uid"]
				}
				if uid == queryUID {
					dataname = name
				}
			}

			ress := b.QueryWorker(queryUID, dataname, source, in.Token, blockQuery.Encode(), in.Method, in.PostForm) //in.QueryRaw
			dataSet[dataname] = ress
		}
	}

	bl.Data = dataSet
	bl.Page = page
	bl.Configuration = conf
	// b.ConfigurationRaw = confRaw
	bl.ConfigurationRaw = tconfiguration
	//bl.Request = r

	// удаляем лишний путь к файлу, добавленную через консоль
	// СЕКЬЮРНО! Если мы вычитаем текущий путь пользователя, то сможем получить доступ к файлам только текущего проекта
	// иначе необходимо будет авторизоваться и правильный путь (например  /console/gui мы не вычтем)
	// НО ПРОБЛЕМА реиспользования ранее загруженных и настроенных путей к шаблонам.
	//tplName = strings.Replace(tplName, Application["client_path"], ".", -1)

	// НЕ СЕКЬЮРНО!
	// вычитаем не текущий client_path а просто две первых секции из адреса к файлу
	// позволяем получить доступ к ранее загруженным путям шаблонов другим пользоватем с другим префиксом
	// ПО-УМОЛЧАНИЮ (для реиспользования модулей и схем)
	sliceMake := strings.Split(tplName, "/")
	if len(sliceMake) < 3 {
		errT := fmt.Errorf("%s", "Error: The path to the module file is incorrect or an error occurred while selecting the module in the block object!")
		//b.ErrorModuleBuild(stat, buildChan, time.Since(t1), errT)
		b.logger.Error(errT)
		return
	}
	tplName = strings.Join(sliceMake[3:], "/")
	tplName = b.cfg.Workingdir + "/" + tplName

	// в режиме отладки пересборка шаблонов происходит при каждом запросе
	var tmpl *template.Template
	if !b.cfg.CompileTemplates.Value {
		if len(tplName) > 0 {
			name := path.Base(tplName)
			if name == "" {
				err = fmt.Errorf("%s","Error: Getting path.Base failed!")
				tmpl = nil
			} else {
				tmpl, _ = template.New(name).Funcs(b.tplfunc.GetFuncMap()).ParseFiles(tplName)
			}
		}
		if &bl != nil && &c != nil {
			if tmpl == nil {
				err = fmt.Errorf("%s","Error: Parsing template file is fail!")
			} else {
				err = tmpl.Execute(&c, bl)
			}
		} else {
			err = fmt.Errorf("%s","Error: Generate data block is fail!")
		}

	} else {
		t.ExecuteTemplate(&c, tplName, b)
	}

	// ошибка при генерации страницы
	if err != nil {
		//b.ErrorModuleBuild(stat, buildChan, time.Since(t1), errT)
		b.logger.Error(err, "Error generated module.")
		return
	}

	blockBody := c.String()

	if tmpl != nil {
		result.Result = template.HTML(blockBody)
	} else {
		result.Result = "<center><h3>Ошибка обработки файла шаблона (файл не найден) при генерации блока.</h3></center>"
	}

	result.Stat = stat
	result.Id = block.Id

	return result
}

// вываливаем ошибку при генерации модуля
func (b *block) ErrorModuleBuild(stat map[string]interface{}, buildChan chan model.ModuleResult, timerRun interface{}, errT error){
	var result model.ModuleResult

	stat["cache"] = "false"
	stat["time"] = timerRun
	result.Stat = stat
	result.Result = template.HTML(fmt.Sprint(errT))
	result.Err = errT

	buildChan <- result

	return
}

// queryUID - ид-запроса
func (b *block) QueryWorker(queryUID, dataname string, source[]map[string]string, token, queryRaw, metod string, postForm url.Values) interface{}  {
	//var resp Response

	resp :=  b.GUIQuery(queryUID, token, queryRaw, metod, postForm)

	//switch x := resp1.(type) {
	//case Response:
	//	resp = resp1.(Response)
	//
	//default:
	//	resp.Data = resp1ч
	//}


	///////////////////////////////////////////
	// Расчет пагенации
	///////////////////////////////////////////

	var m3 model.Response
	b1, _ := json.Marshal(resp)
	json.Unmarshal(b1, &m3)
	var last, current, from, to, size int
	var list []int

	resultLimit := m3.Metrics.ResultLimit
	resultOffset := m3.Metrics.ResultOffset
	size = m3.Metrics.ResultSize

	if size != 0 && resultLimit != 0 {
		j := 0
		for i := 0; i <= size; i = i + resultLimit {
			j ++
			list = append(list, j)
			if i >= resultOffset && i < resultOffset + resultLimit {
				current = j
			}
		}
		last = j
	}

	from = current * resultLimit - resultLimit
	to = from + resultLimit

	// подрезаем список страниц
	lFrom := 0
	if current != 1 {
		lFrom = current - 2
	}
	if lFrom <= 0 {
		lFrom = 0
	}

	lTo := current + 4
	if lTo > last {
		lTo = last
	}
	if lTo <= 0 {
		lTo = 0
	}

	lList := list[lFrom:lTo]

	resp.Metrics = m3.Metrics
	resp.Metrics.PageLast = last
	resp.Metrics.PageCurrent = current
	resp.Metrics.PageList = lList

	resp.Metrics.PageFrom = from
	resp.Metrics.PageTo = to

	///////////////////////////////////////////
	///////////////////////////////////////////

	return resp
}

// вывод ошибки выполнения блока
func (b *block) ErrorPage(err interface{}, w http.ResponseWriter, r *http.Request) {
	p := model.ErrorForm{
		Err: err,
		R:	 *r,
	}
	b.logger.Error(nil, err)

	t := template.Must(template.ParseFiles("./upload/control/templates/errors/500.html"))
	t.Execute(w, p)
}

// вывод ошибки выполнения блока
func (l *block) ModuleError(err interface{}) template.HTML  {
	var c bytes.Buffer

	p := model.ErrorForm{
		Err: err,
	}

	l.logger.Error(nil,err)
	//fmt.Println("ModuleError: ", err)

	wd := l.cfg.Workingdir
	t := template.Must(template.ParseFiles(wd + "/upload/control/templates/errors/503.html"))

	t.Execute(&c, p)
	result := template.HTML(c.String())

	return result
}

// отправка запроса на получения данных из интерфейса GUI
// параметры переданные в строке (r.URL) отправляем в теле запроса
func (b *block) GUIQuery(tquery, token, queryRaw, method string, postForm url.Values) model.Response  {
	var resultInterface interface{}
	var dataResp, returnResp model.Response

	bodyJSON, _ := json.Marshal(postForm)

	// добавляем к пути в запросе переданные в блок параметры ULR-а (возможно там есть параметры для фильтров)
	filters := queryRaw
	if filters != "" {
		filters = "?" + filters
	}

	// ФИКС!
	// добавляем еще токен (cookie) текущего пользователя
	// это нужно для случая, если мы вызываем запрос из запроса и кука не передается
	// а если куки нет, то сбрасывается авторизация
	if token != "" {
		if strings.Contains(filters, "?") {
			filters = filters + "&token=" + token
		} else {
			filters = filters + "?token=" + token
		}
	}

	resultInterface, _ = b.utils.Curl(method, "/query/" + tquery + filters, string(bodyJSON), &dataResp, map[string]string{})

	// нам тут нужен Response, но бывают внешние запросы,
	// поэтому если не Response то дописываем в Data полученное тело
	if dataResp.Data != nil {
		returnResp = dataResp
	} else {
		returnResp.Data = resultInterface
	}

	var dd model.ResponseData
	ff, _ := json.Marshal(dd)
	json.Unmarshal(ff, &dd)

	return returnResp
}


func New(
	cfg config.Config,
	logger log.Log,
	utils utils.Utils,
	function function.Function,
	tplfunc function.TplFunc,
) Block {
	return &block {
		cfg: cfg,
		logger: logger,
		utils: utils,
		function: function,
		tplfunc: tplfunc,
	}	
}

