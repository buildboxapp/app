package lib

import (
	"github.com/buildboxapp/logger"
	"github.com/getlantern/errors"
)

type Lib struct {
	Logger *logger.Log
	State  map[string]string
	UrlAPI string `json:"url_api"`
	UrlGUI string `json:"url_gui"`
}

var StatusCode = RStatus{
	"OK":                       {"Запрос выполнен", 200, "", ""},
	"OKLicenseActivation":      {"Лицензия была активирована", 200, "", ""},
	"Unauthorized":             {"Ошибка авторизации", 401, "", ""},
	"NotCache":                 {"Доступно только в Турбо-режиме", 200, "", ""},
	"NotStatus":                {"Ответ сервера не содержит статус выполнения запроса", 501, "", ""},
	"NotExtended":              {"На сервере отсутствует расширение, которое желает использовать клиент", 501, "", ""},
	"ErrorFormatJson":          {"Ошибка формата JSON-запроса", 500, "ErrorFormatJson", ""},
	"ErrorTransactionFalse":    {"Ошибка выполнения тразакции SQL", 500, "ErrorTransactionFalse", ""},
	"ErrorBeginDB":             {"Ошибка подключения к БД", 500, "ErrorBeginDB", ""},
	"ErrorPrepareSQL":          {"Ошибка подготовки запроса SQL", 500, "ErrorPrepareSQL", ""},
	"ErrorNullParameter":       {"Ошибка! Не передан параметр", 503, "ErrorNullParameter", ""},
	"ErrorQuery":               {"Ошибка запроса на выборку данных", 500, "ErrorQuery", ""},
	"ErrorScanRows":            {"Ошибка переноса данных из запроса в объект", 500, "ErrorScanRows", ""},
	"ErrorNullFields":          {"Не все поля заполнены", 500, "ErrorScanRows", ""},
	"ErrorAccessType":          {"Ошибка доступа к элементу типа", 500, "ErrorAccessType", ""},
	"ErrorGetData":             {"Ошибка доступа данным объекта", 500, "ErrorGetData", ""},
	"ErrorRevElement":          {"Значение было изменено ранее.", 409, "ErrorRevElement", ""},
	"ErrorForbiddenElement":    {"Значение занято другим пользователем.", 403, "ErrorForbiddenElement", ""},
	"ErrorUnprocessableEntity": {"Необрабатываемый экземпляр", 422, "ErrorUnprocessableEntity", ""},
	"ErrorNotFound":            {"Значение не найдено", 404, "ErrorNotFound", ""},
	"ErrorReadConfigDir":       {"Ошибка чтения директории конфигураций", 403, "ErrorReadConfigDir", ""},
	"errorOpenConfigDir":       {"Ошибка открытия директории конфигураций", 403, "errorOpenConfigDir", ""},
	"ErrorReadConfigFile":      {"Ошибка чтения файла конфигураций", 403, "ErrorReadConfigFile", ""},
	"ErrorPortBusy":            {"Указанный порт занят", 403, "ErrorPortBusy", ""},
	"ErrorGone":                {"Объект был удален ранее", 410, "ErrorGone", ""},
	"ErrorShema":               {"Ошибка формата заданной схемы формирования запроса", 410, "ErrorShema", ""},
	"ErrorInitBase":            {"Ошибка инициализации новой базы данных", 410, "ErrorInitBase", ""},
	"ErrorCreateCacheRecord":   {"Ошибка создания объекта в кеше", 410, "ErrorCreateCacheRecord", ""},
	"ErrorUpdateParams":        {"Не переданы параметры для обновления серверов (сервер источник, сервер получатель)", 410, "ErrorUpdateParams", ""},
	"ErrorIntervalProxy":       {"Ошибка переданного интервала (формат: 1000:2000)", 410, "ErrorIntervalProxy", ""},
	"ErrorReservPortProxy":     {"Ошибка выделения порта proxy-сервером", 410, "ErrorReservPortProxy", ""},
}

type RStatus map[string]RestStatus
type RestStatus struct {
	Description string `json:"description"`
	Status      int    `json:"status"`
	Code        string `json:"code"`
	Error       string `json:"error"`
}

type Response struct {
	Data    interface{} `json:"data"`
	Res     interface{} `json:"res"`
	Status  RestStatus  `json:"status"`
	Metrics Metrics     `json:"metrics"`
}

type ResponseData struct {
	Data    []Data      `json:"data"`
	Res     interface{} `json:"res"`
	Status  RestStatus  `json:"status"`
	Metrics Metrics     `json:"metrics"`
}

type Metrics struct {
	ResultSize    int    `json:"result_size"`
	ResultCount   int    `json:"result_count"`
	ResultOffset  int    `json:"result_offset"`
	ResultLimit   int    `json:"result_limit"`
	ResultPage    int    `json:"result_page"`
	TimeExecution string `json:"time_execution"`
	TimeQuery     string `json:"time_query"`
}

type Attribute struct {
	Value  *string `json:"value"`
	Src    *string `json:"src"`
	Tpls   *string `json:"tpls"`
	Status *string `json:"status"`
	Rev    *string `json:"rev"`
	Uuid   *string `json:"uuid"`
}

type Data struct {
	Uid        string               `json:"uid"`
	Id         string               `json:"id"`
	Source     string               `json:"source"`
	Parent     string               `json:"parent"`
	Type       string               `json:"type"`
	Title      string               `json:"title"`
	Rev        string               `json:"rev"`
	Attributes map[string]Attribute `json:"attributes"`
	Linkinid   string               `json:"linkinid"`
	Linkinobj  []Data               `json:"linkinobj"`
}

type Hosts struct {
	Host     string `json:"host"`
	PortFrom int    `json:"portfrom"`
	PortTo   int    `json:"portto"`
	Protocol string `json:"protocol"`
}

// метод, которые проверяем наличие ключа в стейте приложения и если нет, то пишет в лог
func (s *Lib) Get(key string) (value string) {
	value, found := s.State[key]
	if !found {
		err := errors.New("Key '" + key + "' from application state not found")
		s.Logger.Error(err)
	}
	return value
}
