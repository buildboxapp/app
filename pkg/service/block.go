package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/app/pkg/model"
	uuid2 "github.com/google/uuid"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)


// Ping ...
func (s *service) Block(ctx context.Context, in model.ServiceBlockIn) (out model.ServiceBlockOut, err error) {
	var objBlock model.ResponseData
	dataPage 		:= model.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы

	s.utils.Curl("GET", "_objs/"+in.Block, "", &objBlock)

	moduleResult := s.ModuleBuild(objBlock.Data[0], r, dataPage, nil, false)

	w.Write([]byte(moduleResult.Result))

	return r, err
}



// ДЛЯ ПОСЛЕДОВАТЕЛЬНОЙ сборки блока
// получаем объект модуля (отображения)
// p 	- объект переданных в модуль данных блока (запрос/конфигураци)
// r 	- значения реквеста
// page - объект страницы, которую парсим
func (s *service) ModuleBuild(in model.ServicePageIn, block model.Data, page model.Data, values map[string]interface{}, enableCache bool) (result model.ModuleResult) 	{
	var c bytes.Buffer
	var err error


	// заменяем в State localhost на адрес домена (если это подпроцесс то все норм, но если это корневой сервис,
	// то у него url_proxy - localhost и узнать реньше адрес мы не можем, ибо еще домен не инициировался
	// а значит подменяем localhost при каждом обращении к модулю
	if strings.Contains(s.cfg.UrlProxy, "localhost") {
		s.cfg.UrlProxy = "//" + r.Host
	}

	b := model.Block{}
	b.Mx.Lock()
	defer b.Mx.Unlock()

	t1 := time.Now()
	stat := map[string]interface{}{}
	stat["start"] = t1
	stat["status"] = "OK"
	stat["title"] = block.Title
	stat["id"] = block.Id


	// Включаем режим кеширования
	key := ""
	keyParam := ""
	cacheOn, _ := block.Attr("cache", "value")

	ll := s.cfg.BaseCache
	if  ll != "" && cacheOn != "" && enableCache {

		key, keyParam = l.SetCahceKey(r, block)

		// ПРОВЕРКА КЕША (если есть, отдаем из кеша)
		if res, found := l.СacheGet(key, block, r, page, values, keyParam); found {
			stat["cache"] = "true"
			stat["time"] = time.Since(t1)

			result.result = template.HTML(res)
			result.stat = stat

			return result
		}
	}

	b.Value = map[string]interface{}{}

	// обработка всех странных ошибок
	// ВКЛЮЧИТЬ ПОЗЖЕ!!!!
	//defer func() {
	//	if er := recover(); er != nil {
	//		//ft, err := template.ParseFiles("./upload/control/templates/errors/503.html")
	//		//if err != nil {
	//		//	l.Logger.Error(err)
	//		//}
	//		//t = template.Must(ft, nil)
	//
	//		result.result = l.ModuleError(er, r)
	//		result.err = err
	//	}
	//}()

	dataSet		:= make(map[string]interface{})
	dataname 	:= "default" // значение по-умолчанию (будет заменено в потоках)

	tplName, _ 	:= block.Attr("module", "src")
	tquery, _ 	:= block.Attr("tquery", "src")


	// //////////////////////////////////////////////////////////////////////////////
	// в блоке есть настройки поля расширенного фильтра, который можно добавить в самом блоке
	// дополняем параметры request-a, доп. параметрами, которые переданы через блок
	extfilter, _ 	:= block.Attr("extfilter", "value") // дополнительный фильтр для блока
	dv := []model.Data{block}
	extfilter = s.DogParse(extfilter, r, &dv, b.Value)
	extfilter = strings.Replace(extfilter, "?", "", -1)

	// парсим переденную строку фильтра
	m, err := url.ParseQuery(extfilter)
	if err != nil {
		s.logger.Error(err, "Error parsing extfilter from block.")
	}

	// добавляем в URL переданное значение из настроек модуля
	var q url.Values
	for k, v := range m {
		q = r.URL.Query() // Get a copy of the query values.
		q.Add(k, strings.Join(v, ",")) // Add a new value to the set. Переводим обратно в строку из массива
	}
	if len(m) != 0 {
		r.URL.RawQuery = q.Encode() // Encode and assign back to the original query.
	}
	// //////////////////////////////////////////////////////////////////////////////
	// //////////////////////////////////////////////////////////////////////////////


	tconfiguration , _ := block.Attr("configuration", "value")
	tconfiguration = strings.Replace(tconfiguration, "  ", "", -1)


	uuid := uuid2.New()

	if values != nil && len(values) != 0 {
		for k, v := range values {
			if _, found := b.Value[k]; !found {
				b.Value[k] = v
			}
		}
	}

	b.Value["Rand"] =  uuid[1:6]  // переопределяем отдельно для каждого модуля
	b.Value["URL"] = in.Url
	b.Value["Prefix"] = "/" + s.cfg.Domain + "/" +s.cfg.PathTemplates
	b.Value["Domain"] = s.cfg.Domain
	b.Value["CDN"] = s.cfg.UrlFs
	b.Value["Path"] = s.cfg.ClientPath
	b.Value["Title"] = s.cfg.Title
	b.Value["Form"] = in.Form
	b.Value["RequestURI"] = in.RequestURI
	b.Value["Referer"] = in.Referer
	b.Value["Profile"] = in.Profile



	// обработк @-функции в конфигурации
	dv = []model.Data{block}
	dogParseConfiguration := l.DogParse(tconfiguration, r, &dv, b.Value)


	// конфигурация без обработки @-функции
	var confRaw map[string]model.Element
	if tconfiguration != "" {
		err = json.Unmarshal([]byte(tconfiguration), &confRaw)
	}

	// конфигурация с обработкой @-функции
	var conf map[string]model.Element
	if dogParseConfiguration != "" {
		err = json.Unmarshal([]byte(dogParseConfiguration), &conf)
	}

	if err != nil {
		rm, _ := json.Marshal(tconfiguration)
		result.Result = s.ModuleError("Error json-format configurations: " + string(rm), r)
		result.Err = err
		return result
	}

	// сформировал структуру полученных описаний датасетов
	var source []map[string]string
	if d, found := conf["datasets"]; found {
		rm, _ := json.Marshal(d.Source)
		err := json.Unmarshal(rm, &source)
		if err != nil {
			stat["status"] = "error"
			stat["description"] = fmt.Sprint(err)

			result.Result = s.ModuleError(err,r)
			result.Err = err
			result.Stat = stat
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

			ress := l.QueryWorker(queryUID, dataname, source, r)


			//fmt.Println("ress: ", ress)

			dataSet[dataname] = ress
		}

	}



	b.Data = dataSet
	b.Page = page
	b.Configuration = conf
	// b.ConfigurationRaw = confRaw
	b.ConfigurationRaw = tconfiguration
	b.Request = r

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
	tplName = strings.Join(sliceMake[3:], "/")

	tplName = l.State["Workingdir"] + "/" + tplName

	// в режиме отладки пересборка шаблонов происходит при каждом запросе
	var tmpl *template.Template
	if debugMode {
		if len(tplName) > 0 {
			name := path.Base(tplName)
			tmpl, _ = template.New(name).Funcs(FuncMap).ParseFiles(tplName)

		}
		if tmpl != nil {
			tmpl.Execute(&c, b)
		}
	} else {
		t.ExecuteTemplate(&c, tplName, b)
	}

	if tmpl != nil {
		result.result = template.HTML(c.String())
	} else {
		result.result = "<center><h3>Ошибка обработки файла шаблона (файл не найден) при генерации блока.</h3></center>"
	}

	// Включаем режим кеширования
	jj := l.State["BaseCache"]
	if jj != "" && cacheOn != "" && enableCache {
		fmt.Println("кэш включен")
		key, keyParam = l.SetCahceKey(r, block)

		fmt.Println(" Начинаем кешировать")
		// КЕШИРОВАНИЕ данных
		l.CacheSet(key, block, page, c.String(), keyParam)
		// log.Warning("CacheSet: ",fl)
	}

	stat["cache"] = "false"
	stat["time"] = time.Since(t1)
	result.stat = stat


	return result
}

// ДЛЯ ПАРАЛЛЕЛЬНОЙ сборки модуля
// получаем объект модуля (отображения)
func (s *service) ModuleBuildParallel(in model.ServicePageIn, ctxM context.Context, p model.Data, page model.Data, values map[string]interface{}, enableCache bool, buildChan chan model.ModuleResult, wg *sync.WaitGroup) 	{
	defer wg.Done()
	t1 := time.Now()

	result := model.ModuleResult{}

	// проверка на выход по сигналу
	select {
	case <- ctxM.Done():
		return
	default:
	}

	if strings.Contains(s.cfg.UrlProxy, "localhost") {
		s.cfg.UrlProxy = "//" + in.Host
	}

	var c bytes.Buffer
	var b model.Block
	var errT, err error
	var key, keyParam string
	b.Value = map[string]interface{}{}
	result.Id = p.Id

	stat := map[string]interface{}{}
	stat["start"] = t1
	stat["status"] = "OK"
	stat["title"] = p.Title
	stat["id"] = p.Id

	//////////////////////////////
	// Включаем режим кеширования
	//////////////////////////////
	cacheOn, _ := p.Attr("cache", "value")

	if s.cfg.BaseCache != "" && cacheOn != "" && enableCache {

		key, keyParam := s.cache.SetCahceKey(p, in.Path, in.Query)

		// ПРОВЕРКА КЕША (если есть, отдаем из кеша)
		if res, found := l.СacheGet(key, p, r, page, values, keyParam); found {
			stat["cache"] = "true"
			stat["time"] = time.Since(t1)

			result.result = template.HTML(res)
			result.stat = stat

			buildChan <- result
			return
		}
	}
	//////////////////////////////
	//////////////////////////////


	// проверка на выход по сигналу
	select {
	case <- ctxM.Done():
		return
	default:
	}

	// обработка всех странных ошибок
	//defer func() {
	//	if er := recover(); er != nil {
	//		t = template.Must(template.ParseFiles("./upload/control/templates/errors/503.html"))
	//		result.result = ModuleError(er, r)
	//	}
	//}()

	dataSet		:= make(map[string]interface{})
	dataname 	:= "default" // значение по-умолчанию (будет заменено в потоках)

	tplName, _ := p.Attr("module", "src")
	tquery, _ := p.Attr("tquery", "src")


	// //////////////////////////////////////////////////////////////////////////////
	// в блоке есть настройки поля расширенного фильтра, который можно добавить в самом блоке
	// дополняем параметры request-a, доп. параметрами, которые переданы через блок
	extfilter, _ 	:= p.Attr("extfilter", "value") // дополнительный фильтр для блока
	dp := []model.Data{p}
	extfilter = l.DogParse(extfilter, r, &dp, b.Value)
	extfilter = strings.Replace(extfilter, "?", "", -1)

	// парсим переденную строку фильтра
	m, err := url.ParseQuery(extfilter)
	if err != nil {
		l.Logger.Error(err, "Error parsing extfilter from block.")
	}

	// добавляем в URL переданное значение из настроек модуля
	var q url.Values
	for k, v := range m {
		q = r.URL.Query() // Get a copy of the query values.
		q.Add(k, join(v, ",")) // Add a new value to the set. Переводим обратно в строку из массива
	}
	if len(m) != 0 {
		r.URL.RawQuery = q.Encode() // Encode and assign back to the original query.
	}
	// //////////////////////////////////////////////////////////////////////////////
	// //////////////////////////////////////////////////////////////////////////////


	tconfiguration , _ := p.Attr("configuration", "value")
	tconfiguration = strings.Replace(tconfiguration, "  ", "", -1)

	uuid := UUID()

	if values != nil && len(values) != 0 {
		for k, v := range values {
			if _, found := b.Value[k]; !found {
				b.Value[k] = v
			}
		}
	}

	b.Value["Rand"] =  uuid[1:6]  // переопределяем отдельно для каждого модуля
	b.Value["URL"] = r.URL.Query().Encode()
	b.Value["Prefix"] = "/" + Domain + "/" +l.State["PathTemplates"]
	b.Value["Domain"] = Domain
	b.Value["CDN"] = l.State["UrlFs"]
	b.Value["Path"] = ClientPath
	b.Value["Title"] = Title
	b.Value["Form"] = r.Form
	b.Value["RequestURI"] = r.RequestURI
	b.Value["Referer"] = r.Referer()
	b.Value["Profile"] = profile


	// обработк @-функции в конфигурации
	dp = []model.Data{p}
	dogParseConfiguration := l.DogParse(tconfiguration, r, &dp, b.Value)

	// конфигурация без обработки @-функции
	var confRaw map[string]model.Element
	if tconfiguration != "" {
		err = json.Unmarshal([]byte(tconfiguration), &confRaw)
	}

	// конфигурация с обработкой @-функции
	var conf map[string]model.Element
	if dogParseConfiguration != "" {
		err = json.Unmarshal([]byte(dogParseConfiguration), &conf)
	}


	if err != nil {
		result.result = l.ModuleError("Error json-format configurations: "+marshal(tconfiguration), r)
		result.err = err
		buildChan <- result

		//dd := map[string]template.HTML{key:ModuleError("Error json-format configurations: "+marshal(tconfiguration), r)}
		return
	}

	// сформировал структуру полученных описаний датасетов
	var source []map[string]string
	if d, found := conf["datasets"]; found {
		err := json.Unmarshal([]byte(marshal(d.Source)), &source)
		if err != nil {
			result.result = l.ModuleError(err, r)
			buildChan <- result
			return
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

			//fmt.Println("start quert: ")

			ress := l.QueryWorker(queryUID, dataname, source, r)
			//fmt.Println("res query: ", ress)

			dataSet[dataname] = ress
		}

	}



	b.Data = dataSet
	b.Page = page
	b.Metric = model.Metric
	b.Configuration = conf
	//b.ConfigurationRaw = confRaw
	b.ConfigurationRaw = tconfiguration

	b.Request = r

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
		errT = errors.New("Error: The path to the module file is incorrect or an error occurred while selecting the module in the block object!")
		l.ErrorModuleBuild(stat, buildChan, time.Since(t1), errT)
		return
	}
	tplName = strings.Join(sliceMake[3:], "/")
	tplName = l.State["Workingdir"] + "/"+ tplName


	// в режиме отладки пересборка шаблонов происходит при каждом запросе
	if debugMode {
		var tmpl *template.Template
		if len(tplName) > 0 {
			name := path.Base(tplName)
			if name == "" {
				errT = errors.New("Error: Getting path.Base failed!")
				tmpl = nil
			} else {
				tmpl, _ = template.New(name).Funcs(FuncMap).ParseFiles(tplName)
			}

		}


		if &b != nil && &c != nil {
			if tmpl == nil {
				errT = errors.New("Error: Parsing template file is fail!")
			} else {
				errT = tmpl.Execute(&c, b)
			}
		} else {
			errT = errors.New("Error: Generate data block is fail!")
		}

	} else {
		errT = t.ExecuteTemplate(&c, tplName, b)
	}

	// ошибка при генерации страницы
	if errT != nil {
		l.ErrorModuleBuild(stat, buildChan, time.Since(t1), errT)
		l.Logger.Error(errT, "Error generated module.")
		return
	}

	stat["cache"] = "true"
	stat["time"] = time.Since(t1)

	result.result = template.HTML(c.String())
	result.stat = stat



	// Включаем режим кеширования
	if l.State["BaseCache"] != "" && cacheOn != "" && enableCache {
		key, keyParam = l.SetCahceKey(r, p)

		// КЕШИРОВАНИЕ данных
		l.CacheSet(key, p, page, c.String(), keyParam)
	}

	stat["cache"] = "false"
	stat["time"] = time.Since(t1)
	result.stat = stat


	buildChan <- result

	//log.Warning("Stop ", p.Title, "-", time.Since(t1))

	return
}

// вываливаем ошибку при генерации модуля
func (s *service) ErrorModuleBuild(stat map[string]interface{}, buildChan chan ModuleResult, timerRun interface{}, errT error){
	var result ModuleResult

	stat["cache"] = "false"
	stat["time"] = timerRun
	result.stat = stat
	result.result = template.HTML(fmt.Sprint(errT))
	result.err = errT

	buildChan <- result

	return
}

// queryUID - ид-запроса
func (s *service) QueryWorker(queryUID, dataname string, source[]map[string]string, r *http.Request) interface{}  {
	//var resp Response

	resp :=  c.GUIQuery(queryUID, r)

	//switch x := resp1.(type) {
	//case Response:
	//	resp = resp1.(Response)
	//
	//default:
	//	resp.Data = resp1
	//}


	///////////////////////////////////////////
	// Расчет пагенации
	///////////////////////////////////////////

	var m3 Response
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
func (s *service) ErrorPage(err interface{}, w http.ResponseWriter, r *http.Request) {
	p := ErrorForm{
		Err: err,
		R:	 *r,
	}
	log.Error(err)

	t = template.Must(template.ParseFiles("./upload/control/templates/errors/500.html"))
	t.Execute(w, p)
}

// вывод ошибки выполнения блока



func (l *service) ModuleError(err interface{}, r *http.Request) template.HTML  {
	var c bytes.Buffer

	p := ErrorForm{
		Err: err,
		R:	 *r,
	}

	l.Logger.Error(nil,err)
	fmt.Println("ModuleError: ", err)

	wd := l.State["Workingdir"]
	t = template.Must(template.ParseFiles(wd + "/upload/control/templates/errors/503.html"))

	t.Execute(&c, p)
	result = template.HTML(c.String())

	return result
}

// отправка запроса на получения данных из интерфейса GUI
// параметры переданные в строке (r.URL) отправляем в теле запроса
func (c *service) GUIQuery(tquery string, r *http.Request) model.Response  {

	var resultInterface interface{}
	var dataResp, returnResp Response

	formValues := r.PostForm
	bodyJSON, _ := json.Marshal(formValues)

	// добавляем к пути в запросе переданные в блок параметры ULR-а (возможно там есть параметры для фильтров)
	filters := r.URL.RawQuery
	if filters != "" {
		filters = "?" + filters
	}


	// ФИКС!
	// добавляем еще токен (cookie) текущего пользователя
	// это нужно для случая, если мы вызываем запрос из запроса и кука не передается
	// а если куки нет, то сбрасывается авторизация
	cookieCurrent, err := r.Cookie("sessionID")
	token := ""
	if err == nil {
		tokenI := strings.Split(fmt.Sprint(cookieCurrent), "=")
		if len(tokenI) > 1 {
			token = tokenI[1]
		}
		if token != "" {
			if strings.Contains(filters, "?") {
				filters = filters + "&token=" + token
			} else {
				filters = filters + "?token=" + token
			}
		}
	}

	//fmt.Println("filters: ",filters)

	resultInterface, _ = c.Curl(r.Method, "/query/" + tquery + filters, string(bodyJSON), &dataResp)

	//fmt.Println(dataResp)
	//fmt.Println("tquery: ", "/query/" + tquery + filters, "; resultInterface: ", resultInterface)

	// нам тут нужен Response, но бывают внешние запросы,
	// поэтому если не Response то дописываем в Data полученное тело
	if dataResp.Data != nil {
		returnResp = dataResp
	} else {
		returnResp.Data = resultInterface
	}

	var dd ResponseData
	ff, _ := json.Marshal(dd)
	json.Unmarshal(ff, &dd)

	return returnResp
}

