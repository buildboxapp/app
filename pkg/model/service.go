package model

import (
	"html/template"
	"net/http"
	"net/url"
)

type ServiceCacheIn struct {
	Link string `json:"link"`
}

type ServiceIn struct {
	Page string `json:"page"`
	Block string `json:"block"`
	Url string `json:"url"`
	Referer string `json:"referer"`
	RequestURI string `json:"request_uri"`
	Profile ProfileData `json:"profile"`
	Form url.Values `json:"form"`
	Host string `json:"host"`
	Path string `json:"path"`
	Query url.Values `json:"query"`
	QueryRaw string `json:"query_raw"`
	PostForm url.Values `json:"post_form"`
	Token string `json:"token"`
	Method string `json:"method"`

	CachePath string `json:"cache_path"`
	CacheQuery string `json:"cache_url"`

	RequestRaw *http.Request
}

type ServiceBlockOut struct {
	Result template.HTML `json:"result"`
}

type ServicePageOut struct {
	Body string `json:"body"`
}

type AliveOut struct {
	Cache  interface{} `json:"cache"`
 	Config interface{} `json:"config"`
}
