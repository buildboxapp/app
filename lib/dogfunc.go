package app_lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////

type Formula struct {
	Value 		string `json:"value"`
	Document 	[]Data `json:"document"`
	Request		*http.Request
	Inserts		[]*Insert
	Values     	map[string]interface{}	//  параметры переданные в шаблон при генерации страницы (доступны в шаблоне как $.Value)
	App 		*App
}

// Вставка - это одна функция, которая может иметь вложения
// Text - строка вставки, по которому мы будем заменять в общем тексте
type Insert struct {
	Text 		string 		`json:"text"`
	Arguments 	[]string 	`json:"arguments"`
	Result		string 		`json:"result"`
	Functions	Function
}

// Исчисляемая фукнция с аргументами и параметрами
// может иметь вложения
type Function struct {
	Name 		string 		`json:"name"`
	Arguments 	[]string 	`json:"arguments"`
	Result 		string 		`json:"result"`
}

////////////////////////////////////////////////////////////
// !!! ПОКА ТОЛЬКО ПОСЛЕДОВАТЕЛЬНАЯ ОБРАБОТКА (без сложений)
////////////////////////////////////////////////////////////

func (p *Formula) Replace() (result string) {

	p.Parse()
	p.Calculate()

	for _, v := range p.Inserts {
		p.Value = strings.Replace(p.Value, v.Text, v.Result, -1)
	}

	return p.Value
}

func (p *Formula) Parse() bool  {

	if p.Value == "" {
		return false
	}

	//content := []byte(p.Value)
	//pattern := regexp.MustCompile(`@(\w+)\(([\w]+)(?:,\s*([\w]+))*\)`)
	value := p.Value

	pattern := regexp.MustCompile(`@(\w+)\(\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?\)`)
	allIndexes := pattern.FindAllStringSubmatch(value, -1)

	for _, loc := range allIndexes {

		i := Insert{}
		f := Function{}
		i.Functions = f
		p.Inserts = append(p.Inserts, &i)

		strFunc := string(loc[0])

		strFunc1 := strings.Replace(strFunc, "@", "", -1)
		strFunc1 = strings.Replace(strFunc1, ")", "", -1)
		f1 := strings.Split(strFunc1, "(")

		if len(f1) == 1 {	// если не нашли ( значит неверно задана @-фукнций
			return false
		}

		i.Text = strFunc
		i.Functions.Name = f1[0] // название функции

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
			i.Functions.Arguments = argsClear
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
		//		i.Functions.Name = res
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
		//			i.Functions.Arguments = append(i.Functions.Arguments, res)
		//		}
		//	}
		//}
	}

	return true
}

func (p *Formula) Calculate()  {

	for k, v := range p.Inserts {
		param := strings.ToUpper(v.Functions.Name)

		switch param {
		case "RAND":
			uuid := UUID()
			p.Inserts[k].Result = uuid[1:6]
		case "PATH":
			p.Inserts[k].Result = Path(p.Document, v.Functions.Arguments)
		case "REPLACE":
			p.Inserts[k].Result = DReplace(v.Functions.Arguments)
		case "TIME":
			p.Inserts[k].Result = Time(v.Functions.Arguments)
		case "DATEMODIFY":
			p.Inserts[k].Result = DateModify(v.Functions.Arguments)

		case "USER":
			p.Inserts[k].Result = UserObj(p.Request, v.Functions.Arguments)
		case "ROLE":
			p.Inserts[k].Result = UserRole(p.Request, v.Functions.Arguments)
		case "PROFILE":
			p.Inserts[k].Result = UserProfile(p.Request, v.Functions.Arguments)

		case "OBJ":
			p.Inserts[k].Result = Obj(p.Document, v.Functions.Arguments)
		case "URL":
			p.Inserts[k].Result = FuncURL(p.Request, v.Functions.Arguments)

		case "SPLITINDEX":
			p.Inserts[k].Result = SplitIndex(v.Functions.Arguments)

		case "TPLVALUE":
			p.Inserts[k].Result = TplValue(p.Values, v.Functions.Arguments)
		case "CONFIGVALUE":
			p.Inserts[k].Result = ConfigValue(v.Functions.Arguments)
		case "FIELDVALUE":
			p.Inserts[k].Result = FieldValue(p.Document, v.Functions.Arguments)
		case "FIELDSRC":
			p.Inserts[k].Result = FieldSrc(p.Document, v.Functions.Arguments)
		case "FIELDSPLIT":
			p.Inserts[k].Result = FieldSplit(p.Document, v.Functions.Arguments)
		default:
			p.Inserts[k].Result = ""
		}
	}

}


///////////////////////////////////////////////////
// Фукнции @ обработки
///////////////////////////////////////////////////

// Получение значений $.Value шаблона (работает со значением по-умолчанию)
func (c *App) TplValue(v map[string]interface{}, arg []string) (result string) {
	var valueDefault string

	// берем через глобальную переменную, через (c *App) не работает для ф-ций шаблонизатора
	if len(State) == 0 {
		return "Error parsing @-function TplValue (State is null)"
	}

	if len(arg) > 0 {
		param := arg[0]
		if len(arg) == 2 {
			valueDefault = arg[1]
		}

		result, found := v[strings.Trim(param, " ")]
		if !found {
			if valueDefault == "" {
				return "Error parsing @-function TplValue (Value from this key is not found.)"
			}
			return fmt.Sprint(valueDefault)
		}
		return fmt.Sprint(result)

	} else {
		return "Error parsing @-function TplValue (Arguments is null)"
	}

	return fmt.Sprint(result)
}

// Получение значений из конфигурации проекта (хранится State в объекте приложение App)
func (c *App) ConfigValue(arg []string) (result string) {
	var valueDefault string

	// берем через глобальную переменную, через (c *App) не работает для ф-ций шаблонизатора
	if len(State) == 0 {
		return "Error parsing @-function ConfigValue (State is null)"
	}

	if len(arg) > 0 {
		param := arg[0]
		if len(arg) == 2 {
			valueDefault = arg[1]
		}

		result, found := State[strings.Trim(param, " ")]
		if !found {
			if valueDefault == "" {
				return "Error parsing @-function ConfigValue (Value from this key is not found.)"
			}
			return fmt.Sprint(valueDefault)
		}
		return fmt.Sprint(result)

	} else {
		return "Error parsing @-function ConfigValue (Arguments is null)"
	}

	return fmt.Sprint(result)
}

// Получаем значение из разделенной строки по номер
// параметры:
// str - текст (строка)
// sep - разделитель (строка)
// index - порядковый номер (число) (от 0) возвращаемого элемента
// default - значение по-умолчанию (не обязательно)
func (c *App) SplitIndex(arg []string) (result string) {
	var valueDefault string

	if len(arg) > 0 {

		str := Replace(arg[0], "'", "", -1)
		sep := Replace(arg[1], "'", "", -1)
		index := Replace(arg[2], "'", "", -1)
		defaultV := Replace(arg[3], "'", "", -1)

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
func (c *App) Time(arg []string) (result string) {

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
func (c *App) TimeFormat(arg []string) (result string) {
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

		result = timeformat(ss, mask, format)
	}
	if result == "" {
		result = valueDefault
	}

	return result
}

func (c *App) FuncURL(r *http.Request, arg []string) (result string) {
	r.ParseForm()
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
func (c *App) Path(d []Data, arg []string) (result string) {
	var valueDefault string

	if len(arg) > 0 {
		param := strings.ToUpper(arg[0])

		if len(arg) == 2 {
			valueDefault = strings.ToUpper(arg[1])
		}

		if len(State) != 0 {
			switch param {
			case "API":
				result = State["url_api"]
			case "GUI":
				result = State["url_gui"]
			case "PROXY":
				result = State["url_proxy"]
			case "CLIENT":
				result = State["client_path"]
			case "DOMAIN":
				result = State["domain"]
			default:
				result = State["client_path"]
			}
		}
	}

	if result == "" {
		result = valueDefault
	}

	return result
}

// Заменяем значение
func (c *App) DReplace(arg []string) (result string) {
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
func (c *App) UserObj(r *http.Request, arg []string) (result string) {

	//fmt.Println("User")
	//fmt.Println(arg)

	var valueDefault string

	if len(arg) > 0 {

		if len(arg) == 2 {
			valueDefault = strings.ToUpper(arg[1])
		}

		param := strings.ToUpper(arg[0])
		ctxUser := r.Context().Value("User") // текущий профиль пользователя

		var uu ProfileData

		json.Unmarshal([]byte(marshal(ctxUser)), &uu)


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
func (c *App) UserProfile(r *http.Request, arg []string) (result string) {
	if len(arg) > 0 {

		param := strings.ToUpper(arg[0])
		ctxUser := r.Context().Value("User") // текущий профиль пользователя

		var uu *ProfileData
		if ctxUser != nil {
			uu = ctxUser.(*ProfileData)
		}

		if uu != nil {
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

	}
	return result
}

// Получение текущей роли User-а
func (c *App) UserRole(r *http.Request, arg []string) (result string) {
	if len(arg) > 0 {

		param := strings.ToUpper(arg[0])
		param2 := strings.ToUpper(arg[1])
		ctxUser := r.Context().Value("User") // текущий профиль пользователя

		var uu *ProfileData
		if ctxUser != nil {
			uu = ctxUser.(*ProfileData)
		}

		if uu != nil {
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

	}
	return result
}

// Вставляем значения системных полей объекта
func (c *App) Obj(data []Data, arg []string) (result string) {

	d := data[0]
	var valueDefault string
	if len(arg) > 0 {

		param := strings.ToUpper(arg[0])
		if len(arg) == 2 {
			valueDefault = strings.ToUpper(arg[1])
		}

			switch param {
			case "UID":	// получаем все uid-ы из переданного массива объектов
				slRes := []string{}
				for _, v := range data {
					slRes = append(slRes, v.Uid)
				}
				result = strings.Join(slRes, ",")
			case "ID":
				slRes := []string{}
				for _, v := range data {
					slRes = append(slRes, v.Id)
				}
				result = strings.Join(slRes, ",")
			case "SOURCE":
				result = d.Source
			case "TITLE":
				slRes := []string{}
				for _, v := range data {
					slRes = append(slRes, v.Title)
				}
				result = strings.Join(slRes, ",")
			case "TYPE":
				result = d.Type
			default:
				slRes := []string{}
				for _, v := range data {
					slRes = append(slRes, v.Uid)
				}
				result = strings.Join(slRes, ",")
			}
	}
	if result == "" {
		result = valueDefault
	}

	return result
}

// Вставляем значения (Value) элементов из формы
// Если поля нет, то выводит переданное значение (может быть любой символ)
func (c *App) FieldValue(data []Data, arg []string) (result string) {
	var valueDefault string
	d := data[0]

	if len(arg) > 0 {
		param := strings.ToUpper(arg[0])
		if len(arg) == 2 {
			valueDefault = strings.ToUpper(arg[1])
		}

		val, found := d.Attr(param, "value")
		if found {
			result = strings.Trim(val, " ")
		} else {
			result = valueDefault
		}
	}

	return result
}

// Вставляем ID-объекта (SRC) элементов из формы
// Если поля нет, то выводит переданное значение (может быть любой символ)
func (c *App) FieldSrc(data []Data, arg []string) (result string) {
	d := data[0]

	for _, v := range arg {
		val, _ := d.Attr(v, "src")
		result = strings.Trim(val, " ")
	}
	return result
}

// Разбиваем значения по элементу (Value(по-умолчанию)/Src) элементов из формы по разделителю и возвращаем
// значение по указанному номеру (начала от 0)
// Синтаксис: FieldValueSplit(поле, элемент, разделитель, номер_элемента)
// для разделителя есть кодовые слова slash - / (нельзя вставить в фукнцию)
func (c *App) FieldSplit(data []Data, arg []string) (result string) {
	d := data[0]

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

	result = split_v[num]

	return result
}

///////////////////////////////////////////////////
// Фукнции @ обработки наследованные от математического пакета
///////////////////////////////////////////////////

// Добавление даты к переданной
// date - дата, которую модифицируют (значение должно быть в формате времени)
// modificator - модификатор (например "+24h")
// format - формат переданного времени (по-умолчанию - 2006-01-02T15:04:05Z07:00 (формат: time.RFC3339)
func (c *App) DateModify(arg []string) (result string) {

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
	d, err := time.ParseDuration(modificator)
	if err != nil {
		return dateArg
	}

	return fmt.Sprint(date.Add(d))
}




///////////////////////////////////////////////////////////////
// Собачья-обработка (поиск в строке @функций и их обработка)
///////////////////////////////////////////////////////////////
func (c *App) DogParse(p string, r *http.Request, queryData *[]Data, values map[string]interface{}) (result string) {
	s1 := Formula{}

	// прогоняем полученную строку такое кол-во раз, сколько вложенных уровней + 1 (для сравнения)
	for {
		s1.Value = p
		s1.Request = r
		s1.Values = values
		s1.Document = *queryData
		res_parse := s1.Replace()

		if p == res_parse {
			result = res_parse
			break
		}
		p = res_parse
	}

	return
}
///////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////
