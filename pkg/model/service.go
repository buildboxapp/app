package model

import "net/url"

type ServicePageIn struct {
	Page string `json:"page"`
	Url string `json:"url"`
	Referer string `json:"referer"`
	RequestURI string `json:"request_uri"`
	Profile ProfileData `json:"profile"`
	Form url.Values `json:"form"`
	Host string `json:"host"`
	Path string `json:"path"`
	Query url.Values `json:"query"`
}

type ServicePageOut struct {
	Body string `json:"body"`
}

type ServiceBlockIn struct {
	Block string `json:"block"`
}

type ServiceBlockOut struct {
	Body string `json:"body"`
}