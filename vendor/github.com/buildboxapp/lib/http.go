package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// всегде возвращает результат в интерфейс + ошибка (полезно для внешних запросов с неизвестной структурой)
// сериализуем в объект, при передаче ссылки на переменную типа
func Curl(method, urlc, bodyJSON string, response interface{}, headers map[string]string, urlapi, urlgui string) (result interface{}, err error) {
	var mapValues map[string]string
	var req *http.Request
	client := &http.Client{}

	if len(urlapi) > 0 {
		if urlapi[len(urlapi)-1:] != "/" {
			urlapi = urlapi + "/"
		}
	}
	if len(urlgui) > 0 {
		if urlgui[len(urlgui)-1:] != "/" {
			urlgui = urlgui + "/"
		}
	}

	// дополняем путем до API если не передан вызов внешнего запроса через http://
	if urlc == "" {
		urlc = urlapi
	} else {
		if urlc[:4] != "http" {
			if urlc[:1] != "/" {
				urlc = urlapi + urlc
			} else {
				urlc = urlgui + urlc[1:]
			}
		}
	}

	if method == "" {
		method = "POST"
	}

	method = strings.Trim(method, " ")
	values := url.Values{}
	actionType := ""

	//fmt.Println("lib/http.go; Curl; urlc " , urlc, "")
	//fmt.Println("c.State", c.State)
	//c.Logger.Info("lib/http.go; Curl; urlc " , urlc, "; bodyJSON: ", bodyJSON, "; method: ", method)

	// если в гете мы передали еще и json (его добавляем в строку запроса)
	// только если в запросе не указаны передаваемые параметры
	clearUrl := strings.Contains(urlc, "?")

	bodyJSON = strings.Replace(bodyJSON, "  ", "", -1)
	err = json.Unmarshal([]byte(bodyJSON), &mapValues)

	if method == "JSONTOGET" && bodyJSON != "" && clearUrl {
		actionType = "JSONTOGET"
	}
	if method == "JSONTOPOST" && bodyJSON != "" {
		actionType = "JSONTOPOST"
	}

	switch actionType {
	case "JSONTOGET": // преобразуем параметры в json в строку запроса
		if err == nil {
			for k, v := range mapValues {
				values.Set(k, v)
			}
			uri, _ := url.Parse(urlc)
			uri.RawQuery = values.Encode()
			urlc = uri.String()
			req, err = http.NewRequest("GET", urlc, strings.NewReader(bodyJSON))
		} else {
			fmt.Println("Error! Fail parsed bodyJSON from GET Curl: ", err)
		}
	case "JSONTOPOST": // преобразуем параметры в json в тело запроса

		if err == nil {
			for k, v := range mapValues {
				values.Set(k, v)
			}
			req, err = http.NewRequest("POST", urlc, strings.NewReader(values.Encode()))
			req.PostForm = values
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		} else {
			fmt.Println("Error! Fail parsed bodyJSON to POST: ", err)
		}
	default:
		req, err = http.NewRequest(method, urlc, strings.NewReader(bodyJSON))
	}

	if err != nil {
		return "", err
	}

	// дополняем переданными заголовками
	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error request: method:", method, ", url:", urlc, ", bodyJSON:", bodyJSON)
		return "", err
	} else {
		defer resp.Body.Close()
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	responseString := string(responseData)

	// возвращаем объект ответа, если передано - в какой объект класть результат
	if response != nil {
		json.Unmarshal([]byte(responseString), &response)
	}

	// всегда отдаем в интерфейсе результат (полезно, когда внешние запросы или сериализация на клиенте)
	//json.Unmarshal([]byte(responseString), &result)

	return responseString, err
}
