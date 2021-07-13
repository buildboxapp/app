package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"github.com/buildboxapp/app/pkg/i18n"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/lib/log"
)

type utils struct {
	cfg    	model.Config
	logger 	log.Log
	msg  	i18n.I18n
}

type Utils interface {
	AddressProxy() (port string)
	Curl(method, urlc, bodyJSON string, response interface{}, headers map[string]string) (result interface{}, err error)
	RemoveElementFromData(p *model.ResponseData, i int) bool
	DataToIncl(objData []model.Data) []*model.DataTree
	TreeShowIncl(in []*model.DataTree, obj string) (out []*model.DataTree)
	SortItems(p []*model.DataTree, fieldsort string, typesort string)
	Hash(str string) string
	JsonEscape(i string) string
}


func New(cfg model.Config, logger log.Log, msg i18n.I18n) Utils {
	return &utils{
		cfg,
		logger,
		msg,
	}
}

/////////////////////////////////////////////////////
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
/////////////////////////////////////////////////////

// удаляем элемент из слайса
func (u *utils) RemoveElementFromData(p *model.ResponseData, i int) bool {

	if (i < len(p.Data)){
		p.Data = append(p.Data[:i], p.Data[i+1:]...)
	} else {
		//log.Warning("Error! Position invalid (", i, ")")
		return false
	}

	return true
}

func (u *utils) Hash(str string) string {
	if str == "" {
		return ""
	}
	h := sha1.New()
	h.Write([]byte(str))
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	return sha1_hash
}

// экранируем "
// fmt.Println(jsonEscape(`dog "fish" cat`))
// output: dog \"fish\" cat
func (u *utils) JsonEscape(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1:len(s)-1]
}


//func (c *App) GUIQuery(tquery string, r *http.Request) Response  {
//
//	var resultInterface interface{}
//	var dataResp, returnResp Response
//
//	formValues := r.PostForm
//	bodyJSON, _ := json.Marshal(formValues)
//
//	// добавляем к пути в запросе переданные в блок параметры ULR-а (возможно там есть параметры для фильтров)
//	filters := r.URL.RawQuery
//	if filters != "" {
//		filters = "?" + filters
//	}
//
//
//	// ФИКС!
//	// добавляем еще токен (cookie) текущего пользователя
//	// это нужно для случая, если мы вызываем запрос из запроса и кука не передается
//	// а если куки нет, то сбрасывается авторизация
//	cookieCurrent, err := r.Cookie("sessionID")
//	iam := ""
//	if err == nil {
//		tokenI := strings.Split(fmt.Sprint(cookieCurrent), "=")
//		if len(tokenI) > 1 {
//			iam = tokenI[1]
//		}
//		if iam != "" {
//			if strings.Contains(filters, "?") {
//				filters = filters + "&iam=" + iam
//			} else {
//				filters = filters + "?iam=" + iam
//			}
//		}
//	}
//
//	//fmt.Println("filters: ",filters)
//
//	resultInterface, _ = c.Curl(r.Method, "/query/" + tquery + filters, string(bodyJSON), &dataResp)
//
//	//fmt.Println(dataResp)
//	//fmt.Println("tquery: ", "/query/" + tquery + filters, "; resultInterface: ", resultInterface)
//
//	// нам тут нужен Response, но бывают внешние запросы,
//	// поэтому если не Response то дописываем в Data полученное тело
//	if dataResp.Data != nil {
//		returnResp = dataResp
//	} else {
//		returnResp.Data = resultInterface
//	}
//
//	var dd ResponseData
//	ff, _ := json.Marshal(dd)
//	json.Unmarshal(ff, &dd)
//
//	return returnResp
//}
