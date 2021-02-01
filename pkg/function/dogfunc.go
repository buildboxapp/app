package function

import (
	"fmt"
	"github.com/buildboxapp/app/pkg/config"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/lib/log"
	uuid "github.com/satori/go.uuid"
	"regexp"
	"strconv"
	"strings"
	"time"
)


type function struct {
	cfg config.Config
	formula Formula
	dogfunc DogFunc
	tplfunc TplFunc
}

type Function interface {
	Exec(p string, queryData *[]model.Data, values map[string]interface{}, request model.ServiceIn) (result string)
	TplFunc() TplFunc
}

////////////////////////////////////////////////////////////

type formula struct {
	value 		string `json:"value"`
	document 	[]model.Data `json:"document"`
	request		model.ServiceIn
	inserts		[]*insert
	values     	map[string]interface{}	//  параметры переданные в шаблон при генерации страницы (доступны в шаблоне как $.Value)
	cfg 		config.Config
	dogfunc		DogFunc
}

type Formula interface {
	Replace() (result string)
	Parse() bool
	Calculate()
	SetValue(value string)
	SetValues(value map[string]interface{})
	SetDocument(value []model.Data)
	SetRequest(value model.ServiceIn)
	SetInserts(value []*insert)
}

// Вставка - это одна функция, которая может иметь вложения
// Text - строка вставки, по которому мы будем заменять в общем тексте
type insert struct {
	text 		string 		`json:"text"`
	arguments 	[]string 	`json:"arguments"`
	result		string 		`json:"result"`
	dogfuncs	dogfunc
}

// Исчисляемая фукнция с аргументами и параметрами
// может иметь вложения
type dogfunc struct {
	name 		string 		`json:"name"`
	arguments 	[]string 	`json:"arguments"`
	result 		string 		`json:"result"`
	cfg 		config.Config `json:"cfg"`
	tplfunc		TplFunc
}

type DogFunc interface {
	TplValue(v map[string]interface{}, arg []string) (result string)
	ConfigValue(arg []string) (result string)
	SplitIndex(arg []string) (result string)
	Time(arg []string) (result string)
	TimeFormat(arg []string) (result string)
	FuncURL(r model.ServiceIn, arg []string) (result string)
	Path(d []model.Data, arg []string) (result string)
	DReplace(arg []string) (result string)
	UserObj(r model.ServiceIn, arg []string) (result string)
	UserProfile(r model.ServiceIn, arg []string) (result string)
	UserRole(r model.ServiceIn, arg []string) (result string)
	Obj(data []model.Data, arg []string) (result string)
	FieldValue(data []model.Data, arg []string) (result string)
	FieldSrc(data []model.Data, arg []string) (result string)
	FieldSplit(data []model.Data, arg []string) (result string)
	DateModify(arg []string) (result string)
	DogSendmail(arg []string) (result string)
}

////////////////////////////////////////////////////////////
// !!! ПОКА ТОЛЬКО ПОСЛЕДОВАТЕЛЬНАЯ ОБРАБОТКА (без сложений)
////////////////////////////////////////////////////////////

func (p *formula) Replace() (result string) {
	p.Parse()
	p.Calculate()

	for _, v := range p.inserts {
		p.value = strings.Replace(p.value, v.text, v.result, -1)
	}

	return p.value
}

func (p *formula) Parse() bool  {

	if p.value == "" {
		return false
	}

	//content := []byte(p.value)
	//pattern := regexp.MustCompile(`@(\w+)\(([\w]+)(?:,\s*([\w]+))*\)`)
	value := p.value

	pattern := regexp.MustCompile(`@(\w+)\(\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?\)`)
	allIndexes := pattern.FindAllStringSubmatch(value, -1)

	for _, loc := range allIndexes {

		i := insert{}
		f := dogfunc{}
		i.dogfuncs = f
		p.inserts = append(p.inserts, &i)

		strFunc := string(loc[0])

		strFunc1 := strings.Replace(strFunc, "@", "", -1)
		strFunc1 = strings.Replace(strFunc1, ")", "", -1)
		f1 := strings.Split(strFunc1, "(")

		if len(f1) == 1 {	// если не нашли ( значит неверно задана @-фукнций
			return false
		}

		i.text = strFunc
		i.dogfuncs.name = f1[0] // название функции

		// готовим параметры для передачи в функцию обработки
		if len(f1[1]) > 0 {
			arg := f1[1]

			// разбиваем по запятой
			args := strings.Split(arg, ",")

			// очищаем каждый параметр от ' если есть
			argsClear := []string{}
			for _, v := range args{
				v = strings.Trim(v, " ")
				v = strings.Trim(v, "'")
				argsClear = append(argsClear, v)
			}
			i.dogfuncs.arguments = argsClear
		}

		//for j, loc1 := range loc {
		//	res := string(loc1)
		//	if res == "" {
		//		continue
		//	}
		//
		//	// общий текст вставки
		//	if j == 0 {
		//		i.Text = res
		//	}
		//
		//	// название фукнции
		//	if j == 1 {
		//		i.dogfuncs.Name = res
		//	}
		//
		//	// аргументы для функции
		//	if j > 1 {
		//		if strings.Contains(res, "@") {
		//			// рекурсивно парсим вложенную формулу
		//			//argRec := p.Parse(arg)
		//			//f.Arguments = append(f.Arguments, argRec)
		//		} else {
		//			// добавляем аргумент в слайс аргументов
		//			i.dogfuncs.Arguments = append(i.dogfuncs.Arguments, res)
		//		}
		//	}
		//}
	}

	return true
}

func (p *formula) Calculate()  {

	for k, v := range p.inserts {
		param := strings.ToUpper(v.dogfuncs.name)

		switch param {
		case "RAND":
			uuid := uuid.NewV4().String()
			p.inserts[k].result = uuid[1:6]
		case "SENDMAIL":
			p.inserts[k].result = p.dogfunc.DogSendmail(v.dogfuncs.arguments)
		case "PATH":
			p.inserts[k].result = p.dogfunc.Path(p.document, v.dogfuncs.arguments)
		case "REPLACE":
			p.inserts[k].result = p.dogfunc.DReplace(v.dogfuncs.arguments)
		case "TIME":
			p.inserts[k].result = p.dogfunc.Time(v.dogfuncs.arguments)
		case "DATEMODIFY":
			p.inserts[k].result = p.dogfunc.DateModify(v.dogfuncs.arguments)

		case "USER":
			p.inserts[k].result = p.dogfunc.UserObj(p.request, v.dogfuncs.arguments)
		case "ROLE":
			p.inserts[k].result = p.dogfunc.UserRole(p.request, v.dogfuncs.arguments)
		case "PROFILE":
			p.inserts[k].result = p.dogfunc.UserProfile(p.request, v.dogfuncs.arguments)

		case "OBJ":
			p.inserts[k].result = p.dogfunc.Obj(p.document, v.dogfuncs.arguments)
		case "URL":
			p.inserts[k].result = p.dogfunc.FuncURL(p.request, v.dogfuncs.arguments)

		case "SPLITINDEX":
			p.inserts[k].result = p.dogfunc.SplitIndex(v.dogfuncs.arguments)

		case "TPLVALUE":
			p.inserts[k].result = p.dogfunc.TplValue(p.values, v.dogfuncs.arguments)
		case "CONFIGVALUE":
			p.inserts[k].result = p.dogfunc.ConfigValue(v.dogfuncs.arguments)
		case "FIELDVALUE":
			p.inserts[k].result = p.dogfunc.FieldValue(p.document, v.dogfuncs.arguments)
		case "FIELDSRC":
			p.inserts[k].result = p.dogfunc.FieldSrc(p.document, v.dogfuncs.arguments)
		case "FIELDSPLIT":
			p.inserts[k].result = p.dogfunc.FieldSplit(p.document, v.dogfuncs.arguments)
		default:
			p.inserts[k].result = ""
		}
	}

}

func (f *formula) SetValue(value string)  {
	f.value = value
}

func (f *formula) SetValues(value map[string]interface{})  {
	f.values = value
}

func (f *formula) SetDocument(value []model.Data)  {
	f.document = value
}

func (f *formula) SetRequest(value model.ServiceIn)  {
	f.request = value
}

func (f *formula) SetInserts(value []*insert)  {
	f.inserts = value
}


func NewFormula(cfg config.Config, dogfunc DogFunc) Formula {
	return &formula{
		cfg: cfg,
		dogfunc: dogfunc,
	}
}


///////////////////////////////////////////////////
// Фукнции @ обработки
///////////////////////////////////////////////////

// Получение значений $.Value шаблона (работает со значением по-умолчанию)
func (d *dogfunc) TplValue(v map[string]interface{}, arg []string) (result string) {
	var valueDefault string

	if len(arg) > 0 {
		param := arg[0]
		if len(arg) == 2 {
			valueDefault = arg[1]
		}

		result, found := v[strings.Trim(param, " ")]
		if !found {
			if valueDefault == "" {
				return "Error parsing @-dogfunc TplValue (Value from this key is not found.)"
			}
			return fmt.Sprint(valueDefault)
		}
		return fmt.Sprint(result)

	} else {
		return "Error parsing @-dogfunc TplValue (Arguments is null)"
	}

	return fmt.Sprint(result)
}

// Получение значений из конфигурации проекта (хранится State в объекте приложение App)
func (d *dogfunc) ConfigValue(arg []string) (result string) {
	var valueDefault string

	if len(arg) > 0 {
		param := arg[0]
		if len(arg) == 2 {
			valueDefault = arg[1]
		}

		result, err := d.cfg.GetValue(strings.Trim(param, " "))
		if err != nil {
			if valueDefault == "" {
				return "Error parsing @-dogfunc ConfigValue (Value from this key is not found.)"
			}
			return fmt.Sprint(valueDefault)
		}
		return fmt.Sprint(result)

	} else {
		return "Error parsing @-dogfunc ConfigValue (Arguments is null)"
	}

	return fmt.Sprint(result)
}

// Получаем значение из разделенной строки по номер
// параметры:
// str - текст (строка)
// sep - разделитель (строка)
// index - порядковый номер (число) (от 0) возвращаемого элемента
// default - значение по-умолчанию (не обязательно)
func (d *dogfunc) SplitIndex(arg []string) (result string) {
	var valueDefault string

	if len(arg) > 0 {

		str := d.tplfunc.Replace(arg[0], "'", "", -1)
		sep := d.tplfunc.Replace(arg[1], "'", "", -1)
		index := d.tplfunc.Replace(arg[2], "'", "", -1)
		defaultV := d.tplfunc.Replace(arg[3], "'", "", -1)

		in, err := strconv.Atoi(index)
		if err != nil {
			result = "Error! Index must be a number."
		}
		if len(arg) == 4 {
			valueDefault = defaultV
		}

		slice_str := strings.Split(str, sep)
		result = slice_str[in]

		//fmt.Println(str)
		//fmt.Println(sep)
		//fmt.Println(in)
		//fmt.Println(slice_str)
	}
	if result == "" {
		result = valueDefault
	}


	//fmt.Println(result)

	return result
}

// Получение текущей даты
func (d *dogfunc) Time(arg []string) (result string) {

	if len(arg) > 0 {
		param := strings.ToUpper(arg[0])

		switch param {
		case "NOW","THIS":
			result = time.Now().Format("2006-01-02 15:04:05")
		default:
			result = time.Now().String()
		}
	}

	return result
}

// Получение идентификатор User-а
func (d *dogfunc) TimeFormat(arg []string) (result string) {
	var valueDefault string

	if len(arg) > 0 {

		thisdate := strings.ToUpper(arg[0])	// переданное время (строка) можно вручную или Now (текущее)
		mask := strings.ToUpper(arg[1])		// маска для перевода переданного времени в Time
		format := strings.ToUpper(arg[2])	// формат преобразования времени (как вывести)
		if len(arg) == 4 {
			valueDefault = strings.ToUpper(arg[2])
		}

		ss := thisdate
		switch thisdate {
		case "NOW":
			ss = time.Now().UTC().String()
			mask = "2006-01-02 15:04:05"
		}

		result = d.tplfunc.Timeformat(ss, mask, format)
	}
	if result == "" {
		result = valueDefault
	}

	return result
}

func (d *dogfunc) FuncURL(r model.ServiceIn, arg []string) (result string) {
	var valueDefault string

	if len(arg) > 0 {
		param := arg[0]
		result = strings.Join(r.Form[param], ",")

		if len(arg) == 2 {
			valueDefault = arg[1]
		}
	}

	if result == "" {
		result = valueDefault
	}


	return result
}

// Вставляем значения системных полей объекта
func (d *dogfunc) Path(dm []model.Data, arg []string) (result string) {
	var valueDefault string

	if len(arg) > 0 {
		param := strings.ToUpper(arg[0])

		if len(arg) == 2 {
			valueDefault = strings.ToUpper(arg[1])
		}

			switch param {
			case "API":
				result = d.cfg.UrlApi
			case "GUI":
				result = d.cfg.UrlGui
			case "PROXY":
				result = d.cfg.UrlProxy
			case "CLIENT":
				result = d.cfg.ClientPath
			case "DOMAIN":
				result = d.cfg.Domain
			default:
				result = d.cfg.ClientPath
			}
	}

	if result == "" {
		result = valueDefault
	}

	return result
}

// Заменяем значение
func (d *dogfunc) DReplace(arg []string) (result string) {
	var count int
	var str, oldS, newS string
	var err error

	if len(arg) > 0 {
		str = arg[0]
		oldS = arg[1]
		newS = arg[2]

		if len(arg) >= 4 {
			countString := arg[3]
			count, err = strconv.Atoi(countString)
			if err != nil {
				count = -1
			}
		}
		result = strings.Replace(str, oldS, newS, count)
	}

	return result
}

// Получение идентификатор User-а (для Cockpit-a)
func (d *dogfunc) UserObj(r model.ServiceIn, arg []string) (result string) {

	//fmt.Println("User")
	//fmt.Println(arg)

	var valueDefault string

	if len(arg) > 0 {

		if len(arg) == 2 {
			valueDefault = strings.ToUpper(arg[1])
		}

		param := strings.ToUpper(arg[0])
		uu := r.Profile // текущий профиль пользователя

		if &uu != nil {
			switch param {
			case "UID","ID":
				result = uu.Uid
			case "PHOTO":
				result = uu.Photo
			case "AGE":
				result = uu.Age
			case "NAME":
				result = uu.First_name + " " + uu.Last_name
			case "EMAIL":
				result = uu.Email
			case "STATUS":
				result = uu.Status
			default:
				result = uu.Uid
			}
		}

	}
	if result == "" {
		result = valueDefault
	}

	return result
}

// Получение UserProfile (для Cockpit-a)
func (d *dogfunc) UserProfile(r model.ServiceIn, arg []string) (result string) {
	if len(arg) > 0 {

		param := strings.ToUpper(arg[0])

		var uu = r.Profile
		role := uu.CurrentRole

		switch param {
		case "UID","ID":
			result = role.Uid
		case "TITLE":
			result = role.Title
		case "DEFAULT":
			result, _ = role.Attr("profile_default", "value")
		default:
			result = uu.Uid
		}


	}
	return result
}

// Получение текущей роли User-а
func (d *dogfunc) UserRole(r model.ServiceIn, arg []string) (result string) {
	if len(arg) > 0 {

		param := strings.ToUpper(arg[0])
		param2 := strings.ToUpper(arg[1])

		var uu = r.Profile
		role := uu.CurrentRole

		switch param {
		case "UID","ID":
			result = role.Uid
		case "TITLE":
			result = role.Title
		case "ADMIN":
			result, _ = role.Attr("role_default", "value")
		case "HOMEPAGE":
			if param2 == "SRC" {
				result, _ = role.Attr("homepage", "src")
			} else {
				result, _ = role.Attr("homepage", "value")
			}
		case "DEFAULT":
			result, _ = role.Attr("default", "value")
		default:
			result = uu.Uid
		}

	}
	return result
}

// Вставляем значения системных полей объекта
func (d *dogfunc) Obj(data []model.Data, arg []string) (result string) {
	var valueDefault, r string
	var res = []string{}
	if len(arg) == 0 {
		result = "Ошибка в переданных параметрах"
	}

	param := strings.ToUpper(arg[0])
	separator := ","	// значение разделителя по-умолчанию

	if len(arg) == 0 {
		return "Ошибка в переданных параметрах."
	}
	if len(arg) == 2 {
		valueDefault = arg[1]
	}
	if len(arg) == 3 {
		separator = arg[2]
	}

	for _, d := range data {
		switch param {
		case "UID":	// получаем все uid-ы из переданного массива объектов
			r = d.Uid
		case "ID":
			r = d.Id
		case "SOURCE":
			r = d.Source
		case "TITLE":
			r = d.Title
		case "TYPE":
			r = d.Type
		default:
			r = d.Uid
		}
		res = append(res, r)
	}
	result = d.tplfunc.Join(res, separator)

	if result == "" {
		result = valueDefault
	}

	return result
}

// Вставляем значения (Value) элементов из формы
// Если поля нет, то выводит переданное значение (может быть любой символ)
func (d *dogfunc) FieldValue(data []model.Data, arg []string) (result string) {
	var valueDefault, separator string
	var resSlice = []string{}

	separator = ","	// значение разделителя по-умолчанию

	if len(arg) == 0 {
		return "Ошибка в переданных параметрах."
	}

	param := arg[0]
	if len(arg) == 2 {
		valueDefault = arg[1]
	}
	if len(arg) == 3 {
		separator = arg[2]
	}

	for _, d := range data {
		val, found := d.Attr(param, "value")
		if found {
			resSlice = append(resSlice, strings.Trim(val, " "))
		}
	}
	result = d.tplfunc.Join(resSlice, separator)

	if result == "" {
		result = valueDefault
	}

	return result
}

// Вставляем ID-объекта (SRC) элементов из формы
// Если поля нет, то выводит переданное значение (может быть любой символ)
func (d *dogfunc) FieldSrc(data []model.Data, arg []string) (result string) {
	var valueDefault, separator string
	var resSlice = []string{}

	if len(arg) == 0 {
		return "Ошибка в переданных параметрах."
	}

	param := arg[0]
	if len(arg) == 2 {
		valueDefault = arg[1]
	}
	if len(arg) == 3 {
		separator = arg[2]
	}

	for _, d := range data {
		val, found := d.Attr(param, "src")
		if found {
			resSlice = append(resSlice, strings.Trim(val, " "))
		}
	}
	result = d.tplfunc.Join(resSlice, separator)

	if result == "" {
		result = valueDefault
	}

	return result
}

// Разбиваем значения по элементу (Value(по-умолчанию)/Src) элементов из формы по разделителю и возвращаем
// значение по указанному номеру (начала от 0)
// Синтаксис: FieldValueSplit(поле, элемент, разделитель, номер_элемента)
// для разделителя есть кодовые слова slash - / (нельзя вставить в фукнцию)
func (d *dogfunc) FieldSplit(data []model.Data, arg []string) (result string) {
	var resSlice = []string{}
	var r string

	if len(arg) == 0 {
		return "Ошибка в переданных параметрах."
	}

	if len(arg) < 4 {
		return "Error! Count params must have 4 (field, element, separator, number)"
	}
	field := arg[0]
	element := arg[1]
	sep := arg[2]
	num_str := arg[3]

	if element == "" {
		element = "value"
	}

	// 1. преобразовали в номер
	num, err := strconv.Atoi(num_str)
	if err != nil {
		return fmt.Sprint(err)
	}

	for _, d := range data {
		// 2. получили значение поля
		val, found := d.Attr(field, element)

		if !found {
			return "Error! This field is not found."
		}
		in := strings.Trim(val, " ")
		if sep == "slash" {
			sep = "/"
		}

		// 3. разделили и получили нужный элемент
		split_v := strings.Split(in, sep)
		if len(split_v) < num {
			return "Error! Array size is less than the passed number"
		}

		r = split_v[num]
		resSlice = append(resSlice, r)
	}

	result = d.tplfunc.Join(resSlice, ",")

	return result
}

// Добавление даты к переданной
// date - дата, которую модифицируют (значение должно быть в формате времени)
// modificator - модификатор (например "+24h")
// format - формат переданного времени (по-умолчанию - 2006-01-02T15:04:05Z07:00 (формат: time.RFC3339)
func (d *dogfunc) DateModify(arg []string) (result string) {

	if len(arg) < 2 {
		return "Error! Count params must have min 2 (date, modificator; option: format)"
	}
	dateArg := arg[0]
	modificator := arg[1]

	format := "2006-01-02 15:04:05"
	if len(arg) == 3 {
		format = arg[2]
	}

	// преобразуем полученную дату из строки в дату
	date, err := time.Parse(format, dateArg)
	if err != nil {
		fmt.Println("err: ", err)
		return dateArg
	}

	// преобразуем модификатор во время
	p, err := time.ParseDuration(modificator)
	if err != nil {
		return dateArg
	}

	return fmt.Sprint(date.Add(p))
}


///////////////////////////////////////////////////////////////
// Отправляем почтового сообщения
func (d *dogfunc) DogSendmail(arg []string) (result string) {
	if len(arg) < 9 {
		return "Error! Count params must have min 9 (server, port, user, pass, from, to, subject, message, turbo: string)"
	}
	result = d.tplfunc.Sendmail(arg[0], arg[1], arg[2], arg[3], arg[4], arg[5], arg[6], arg[7], arg[8])

	return result
}

func NewDogFunc(cfg config.Config, tplfunc TplFunc) DogFunc {
	return &dogfunc{
		cfg: cfg,
		tplfunc: tplfunc,
	}
}


///////////////////////////////////////////////////////////////
// Собачья-обработка (поиск в строке @функций и их обработка)
///////////////////////////////////////////////////////////////
func (d *function) Exec(p string, queryData *[]model.Data, values map[string]interface{}, request model.ServiceIn) (result string) {

	// прогоняем полученную строку такое кол-во раз, сколько вложенных уровней + 1 (для сравнения)
	for {
		d.formula.SetValue(p)
		d.formula.SetValues(values)
		d.formula.SetDocument(*queryData)
		d.formula.SetRequest(request)
		res_parse := d.formula.Replace()

		if p == res_parse {
			result = res_parse
			break
		}
		p = res_parse
	}

	return
}

func (d *function) TplFunc() TplFunc {
	return d.tplfunc
}

func New(cfg config.Config, logger log.Log) Function {
	tplfunc := NewTplFunc(cfg, logger)
	dogfunc := NewDogFunc(cfg, tplfunc)
	formula := NewFormula(cfg, dogfunc)

	return &function{
		cfg: cfg,
		formula: formula,
		dogfunc: dogfunc,
		tplfunc: tplfunc,
	}
}