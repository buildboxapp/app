package handlers

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

// обработчик запроса
func JSQuery(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	qattr := QueryAttribute{}
	queryID := vars["obj"]

	// распарсили запрос в r.Form (все типы запросов)
	r.ParseMultipartForm(defaultMaxMemory)

	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	/////////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////////
	// при запросе подзапроса RUN на стононе API происходит вызов из АПИ запроса на ГУИ, что само по себе плохо
	// однако задача в том, чтобы передать во внешний запрос текущего состояния r
	// транспорт единственный - body в NewRequest (другие параметры r.Form не передаются (см.код)
	// поэтому используем метод OPTION - он сигнализирует, что надо заменить r.Form переданными в body запроса значениями типа url.Values

	// КАСТОМ - ПЕРЕДЕЛАТЬ
	// если вызван запрос через метод OPTION - дополняем содержимое r.Form переданными значениями в BODY

	/////////////////////////////////////////////////////////////////////////////////
	// Всегда дополняем r значениями переданными в теле
	var formValue url.Values
	var flagJson = false
	if len(body) != 0 {
		// разбираем тело запроса
		if err := json.Unmarshal([]byte(body), &formValue); err != nil {
			errStr := "Ошибка преобразования запроса в url.Values => " + string(body)
			fmt.Errorf("%s", errStr)
		} else {
			if len(formValue) > 0 {
				r.Form = formValue
			}
			flagJson = true
		}
	}


	/////////////////////////////////////////////////////////////////////////////////
	// Дополняем r значениями из урла запроса
	vaURL := r.URL.Query()


	// ФИКС! Если приходит GET то r.Form пустое и имеет неверный формат. Делаем правильный
	if r.Method == "GET" {
		r.Form = r.PostForm
	}

	if len(vaURL) > 0 {
		if !flagJson {
			r.Form = vaURL
		} else {
			// дополняем текущий r.Form значениями из r.URL.Query
			for keyU, valueU := range vaURL {

				if len(valueU) > 0 {
					for _, v1 := range valueU {

						if v2, found := r.Form[keyU]; found { // если в форме уже есть этот ключ, то пробегаем его значения и если нет, то добавляем
							// если значение из урл-а есть уже в значениях Форм, то не добавляем
							flagFormI := false
							for _, valueForm := range v2 {
								if valueForm == v1 {
									flagFormI = true
								}
							}

							// значекния нет - добавляем
							if !flagFormI {
								r.Form.Add(keyU, v1)
							}

						} else {
							r.Form.Add(keyU, v1)
						}
					}
				}
			}
		}
	}

	resp, queryUID, _ := GetQuery(queryID, r, &qattr)

	var dataObjs ResponseData
	b1, _ := json.Marshal(resp)
	json.Unmarshal(b1, &dataObjs)

	// ФИКС!!! проверка на сам запрос получения триггеров
	// если не сделать, то происходит зацикливание
	if queryID != "query_tpltriggers" {

		/////////////////   ОБРАБОТКА ТРИГГЕРА НА GET ОБЪЕКТОВ ЗАПРОСА   /////////////////
		TriggerRun(dataObjs.Data, r, "get", "after", "")
		/////////////////////////////////////////////////////////////////////////////////

		/////////////////   ОБРАБОТКА ТРИГГЕРА НА ВЫПОЛНЕНИЕ ЗАПРОСА   /////////////////
		// в качестве filter_query передаем uid текущего запроса, значит мы обрабатываем его результаты в триггере
		TriggerRun(dataObjs.Data, r, "", "after", queryUID)
		/////////////////////////////////////////////////////////////////////////////////
	}


	//log.Info("Ответ на запрос: ", queryID, "; данные: ", b1)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(b1)
}
