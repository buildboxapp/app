package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/buildboxapp/app/pkg/model"
	"html/template"
	"strings"
	"sync"
	"time"
)


// Page ...
func (s *service) Page(ctx context.Context, in model.ServiceIn) (out model.ServicePageOut, err error) {
	var objPages, objPage model.ResponseData

	// ПЕРЕДЕЛАТЬ или на кеширование страниц и на доп.проверку
	if in.Page == "" {
		// получаем все страницы текущего приложения
		s.utils.Curl("GET", "_link?obj="+s.cfg.DataSource+"&source="+s.cfg.TplAppPagesPointsrc+"&mode=out", "", &objPages, map[string]string{})

		fmt.Println(objPages)
		for _, v := range objPages.Data {
			if def, _ := v.Attr("default", "value"); def == "checked" {
				in.Page = v.Uid
			}
		}
	}
	if in.Page == "" {
		ff, _ := json.Marshal(objPages)
		err = fmt.Errorf("%s", "Error: not default page (" + fmt.Sprint(ff) + ")")
		return out, err
	}

	// запрос объекта страницы
	_, err = s.utils.Curl("GET", "_objs/"+in.Page, "", &objPage, map[string]string{})
	if err != nil {
		err = fmt.Errorf("%s (%s)", "Error: Fail GET-request!", err)
		return out, err
	}

	//fmt.Println("objPage: ", objPage)

	// ФИКС! иногда в разных приложениях называют одинаково страницы.
	// удаляем из объекта objPage значения не текущего приложения
	if len(objPage.Data) > 1 {
		for k, v := range objPage.Data {
			app, _ := v.Attr("app", "src")
			if app != s.cfg.DataUid {
				s.utils.RemoveElementFromData(&objPage, k)
			}
		}
	}


	// формируем значение переменных, переданных в страницу
	values := map[string]interface{}{}


	values["Prefix"] = s.cfg.ClientPath + s.cfg.PathTemplates
	values["Domain"] = s.cfg.Domain
	values["Path"] = s.cfg.ClientPath
	values["CDN"] = ""
	values["Title"] = s.cfg.Title
	values["URL"] = in.Url
	values["Referer"] = in.Referer
	values["RequestURI"] = in.RequestURI
	values["Profile"] = in.Profile

	out.Body = s.BPage(in, objPage, values)

	return out, err
}


// Собираем страницу
func (s *service) BPage(in model.ServiceIn, objPage model.ResponseData, values map[string]interface{}) string {

	var objMaket, objBlocks model.ResponseData
	var t *template.Template
	moduleResult := model.ModuleResult{}
	statModule := map[string]interface{}{}

	// флаг режима генерации модулей (последовательно/параллельно)
	p := &model.Page{}
	p.Title 	= s.cfg.Title
	p.Domain 	= s.cfg.Domain
	p.Metric	= template.HTML(s.cfg.Metric)
	p.Prefix 	= values["Prefix"]
	//p.Request 	= values["Request"]
	p.CSS 		= []string{}
	p.JS 		= []string{}
	p.JSH 		= []string{}
	p.CSSC		= []string{}
	p.JSC		= []string{}
	p.Stat		= []interface{}{}
	p.Blocks 	= map[string]interface{}{}

	if len(objPage.Data) == 0 {
		return "Error: Object page is null."
	}

	pageUID := objPage.Data[0].Uid
	maketUID, _ := objPage.Data[0].Attr("maket", "src")


	// 1.0 проверка на принадлежность страницы текущему проекту
	// ДОДЕЛАТЬ СРОЧНО!!!

	// 2 запрос на объекты блоков страницы
	s.utils.Curl("GET", "_link?obj="+pageUID+"&source="+s.cfg.TplAppBlocksPointsrc+"&mode=in", "", &objBlocks, map[string]string{})

	//for _, v := range objBlocks.Data {
	//	fmt.Println("objBlocks: ", v.Title, v.Id)
	//}

	// 3 запрос на объект макета
	s.utils.Curl("GET", "_objs/"+maketUID, "", &objMaket, map[string]string{})

	// 4 из объекта макета берем путь к шаблону + css и js
	maketFile, _ := objMaket.Data[0].Attr("file", "value")
	maketCSS, _ := objMaket.Data[0].Attr("css", "value")
	maketJS, _ := objMaket.Data[0].Attr("js", "value")
	maketJSH, _ := objMaket.Data[0].Attr("jsh", "value")
	maketJSC, _ := objMaket.Data[0].Attr("js_custom", "value")
	maketCSSC, _ := objMaket.Data[0].Attr("css_custom", "value")

	// 5 добавляем в объект страницы список файлов css и js
	for _, v := range strings.Split(maketCSS, ";") {
		p.CSS = append(p.CSS, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(maketJS, ";") {
		p.JS = append(p.JS, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(maketJSH, ";") {
		p.JSH = append(p.JSH, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(maketJSC, ";") {
		p.JSC = append(p.JSC, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(maketCSSC, ";") {
		p.CSSC = append(p.CSSC, strings.TrimSpace(v))
	}


	// 3 сохраняем схему
	var i interface{}
	shemaJSON, _ := objPage.Data[0].Attr("shema", "value")
	json.Unmarshal([]byte(shemaJSON), &i)
	if i == nil {
		return "Error! Fail json shema!"
	}
	p.Shema = i

	// 4 запускаем сборку модулей (получаем сгенерированный template.HTML без JS и CSS
	// шаблоны рендерятся в каждом модуле отдельно (можно далее хранить в кеше)

	if s.cfg.BuildModuleParallel.Value {
		ctx := context.WithValue(context.Background(), "timeout", s.cfg.TimeoutBlockGenerate.Value)
		ctx, cancel := context.WithCancel(ctx)

		// ПАРАЛЛЕЛЬНО
		wg := &sync.WaitGroup{}
		var buildChan = make(chan model.ModuleResult, len(objBlocks.Data))

		for _, v := range objBlocks.Data {
			idBlock, _ := v.Attr("id", "value") 	// название блока

			if strings.Contains(shemaJSON, idBlock) {		// наличие этого блока в схеме
				wg.Add(1)
				go s.block.ModuleBuildParallel(in, ctx, v, objPage.Data[0], values, true,  buildChan, wg)
			}
		}

		// ждем завершения интервала и вызываем завершение контекста для запущенных воркеров
		exitTimer := make(chan struct{})
		timerBlockGen := time.NewTimer(s.cfg.TimeoutBlockGenerate.Value)
		flagWG := true
		go func() {
			select {
			case <- timerBlockGen.C:
				flagWG = false
				cancel()
				return
			case <- exitTimer:
				timerBlockGen.Stop()
				return
			}
		}()

		// отменяем ожидание wg при условии, что завершился таймаут и нам не нужны результаты недополученных ModuleBuildParallel
		// wg завершатся сами через defer позже
		if flagWG {
			wg.Wait()
		}
		if timerBlockGen.Stop() {
			exitTimer <- struct{}{}
		}

		close(buildChan)

		for k := range buildChan {
			p.Blocks[k.Id] = k.Result
			p.Stat = append(p.Stat, k.Stat)
		}

	} else {

		// ПОСЛЕДОВАТЕЛЬНО
		for _, v := range objBlocks.Data {

			idBlock, _ := v.Attr("id", "value") 	// название блока
			if strings.Contains(shemaJSON, idBlock) {		// наличие этого блока в схеме
				moduleResult = s.block.Generate(in, v, objPage.Data[0], values, true)

				p.Blocks[v.Id] = moduleResult.Result
				statModule = moduleResult.Stat

				statModule["id"] = v.Id
				statModule["title"] = v.Title
				p.Stat = append(p.Stat, statModule)
			}
		}

	}

	//fmt.Println("Statistic generate page: ", p.Stat)
	//log.Warning("Time всего: ", time.Since(t1))


	// 5 генерируем страницу, использую шаблон выбранной в объекте страницы, схему
	var c bytes.Buffer

	// СЕКЬЮРНО! Если мы вычитаем текущий путь пользователя, то сможем получить доступ к файлам только текущего проекта
	// иначе необходимо будет авторизоваться и правильный путь (например  /console/gui мы не вычтем)
	// НО ПРОБЛЕМА реиспользования ранее загруженных и настроенных путей к шаблонам.
	//maketFile = strings.Replace(maketFile, Application["client_path"], ".", -1)

	// НЕ СЕКЬЮРНО!
	// вычитаем не текущий client_path а просто две первых секции из адреса к файлу
	// позволяем получить доступ к ранее загруженным путям шаблонов другим пользоватем с другим префиксом
	// ПО-УМОЛЧАНИЮ (для реиспользования модулей и схем)
	sliceMake := strings.Split(maketFile, "/")
	maketFile = strings.Join(sliceMake[3:], "/")

	maketFile = s.cfg.Workingdir + "/"+ maketFile

	// в режиме отладки пересборка шаблонов происходит при каждом запросе
	if s.cfg.CompileTemplates.Value {
		//t = template.Must(template.New(maketFile).Funcs(funcMap).ParseFiles(maketFile))
		t = template.Must(template.ParseFiles(maketFile))
		t.Execute(&c, p)
	} else {
		t.ExecuteTemplate(&c, maketFile, p)
	}


	return c.String()
}


// возвращаем сформированную страницу в template.HTML (для cockpit-a и dashboard)
//func (s *service) TIndex(w http.ResponseWriter, r *http.Request, Config map[string]string) template.HTML {
//
//	var objPage, objApp model.ResponseData
//	vars := mux.Vars(r)
//	page := vars["obj"] // ид-страницы передается через переменную obj
//
//	// указатель на профиль текущего пользователя
//	ctx := r.Context()
//	var profile model.ProfileData
//	profileRaw := ctx.Value("UserRaw")
//	json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)
//
//
//	// можем задать также через &page=страница
//	if r.FormValue("page") != "" {
//		page = r.FormValue("page")
//	}
//
//	if page == "" {
//		return ""
//	}
//
//	// заменяем значения при вызове ф-ции из GUI ибо они пустые, ведь приложение полностью не инициализировано через конфиг
//
//	if page == "" {
//		return template.HTML("Error: Not id page")
//	}
//
//	// запрос объекта страницы
//	s.utils.Curl("GET", "_objs/"+page, "", &objPage)
//
//	//fmt.Println("objPage: ", objPage)
//
//	if &objPage == nil {
//		return template.HTML("Error: Not found page-object.") // если не найден объект страницы
//	}
//
//	if len(objPage.Data) == 0 {
//		return template.HTML("Error: Not found page-object.") // если не найден объект страницы
//	}
//
//	// Uid-приложения
//	appUid, found := objPage.Data[0].Attr("app", "src")
//	if !found {
//		return template.HTML("Error: Not selected application from this page.")
//	}
//
//	// запрос объекта приложения
//	s.utils.Curl("GET", "_objs/"+appUid, "", &objApp)
//	if &objApp == nil {
//		return template.HTML("Error: Not found application-object.") // если не найден объект приложения
//	}
//
//	//fmt.Println("objApp: ", objApp)
//
//	// получаем значения аттрибутов для данного приложения
//	path_template, found := objApp.Data[0].Attr("path_templates", "value")
//	if !found {
//		return template.HTML("Error: Not selected path_templates from this application.")
//	}
//
//	// получаем значения аттрибутов для данного приложения
//	tpl_app_blocks_pointsrc, found := objApp.Data[0].Attr("tpl_app_blocks", "src")
//	if !found {
//		return template.HTML("Error: Not selected tpl_app_blocks from this application.")
//	}
//
//	//pp := strings.Split(Domain, "/")
//	//if len(pp) == 1 {
//	//	ClientPath = Domain + "/" + "gui"
//	//}
//
//	// получили значение Request в json - возвращаем в http.Request
//	//var PageRequest *http.Request
//	//json.Unmarshal([]byte(JRequest), &PageRequest)
//
//	// формируем значение переменных, переданных в страницу
//	values := map[string]interface{}{}
//	values["Prefix"] = s.cfg.ClientPath + path_template
//	values["Domain"] = s.cfg.Domain
//	values["Path"] = s.cfg.ClientPath
//	values["CDN"] = ""
//	values["Title"] = s.cfg.Title
//	values["URL"] = r.URL.Query().Encode()
//	values["Referer"] = r.Referer()
//	values["RequestURI"] = r.RequestURI
//	values["Profile"] = profile
//
//
//	result := s.BPage(in, tpl_app_blocks_pointsrc, objPage, values)
//
//	return template.HTML(result)
//}


// генерируем один блок через внутренний запрос - для cocpit-a
//func (s *service) TBlock(r *http.Request, block model.Data, Config map[string]string) template.HTML {
//	dataPage 		:= model.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы
//	moduleResult := s.ModuleBuild(block, r, dataPage, nil, false)
//
//	return moduleResult.Result
//}


// Параметры обязательные для задания
// Удаление кешей независимо от контекста текущего процесса (подключаемся к новому неймспейсу)
// &namespace - таблица в reindexer
// &link - связи для выборки (фиксируем uid-страницы и uid-блока) (может быть значение all - удалить все значения кеша)
//func ClearCache(w http.ResponseWriter, r *http.Request) {
//
//	var err error
//	var countDeleted int
//	status := "OK"
//	ns 		:= r.FormValue("namespace")
//	link 	:= r.FormValue("link")
//
//	if ns == "" || link == "" {
//		ResponseJSON(w, "Parametrs: &namespace=, &link=", "ErrorNullParameter", err, nil)
//		return
//	}
//
//	ns = strings.Replace(ns, "/", "_", -1) //заменяем для имен приложений из ru/ru в формат ru_ru
//	if ns == "" {
//		ns = Namespace
//	}
//
//	DBCache_clear := reindexer.NewReindex(BaseCache)
//	err = DBCache_clear.OpenNamespace(ns, reindexer.DefaultNamespaceOptions(), Value{})
//
//	if link == "all" {
//		// паременты не переданы - удаляем все объекты в заданном неймспейсе
//		countDeleted, err = DBCache_clear.Query(ns).
//			Not().WhereString("Uid", reindexer.EQ, "").Delete()
//	} else {
//		// паременты не переданы - удаляем согласно шаблону
//		countDeleted, err = DBCache_clear.Query(ns).
//			Where("Link", reindexer.SET, link).Delete()
//	}
//
//
//	ResponseJSON(w,  countDeleted, status, err, nil)
//}
//
