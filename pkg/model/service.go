package model

type ServicePageIn struct {
	Page string `json:"page"`
	Url string `json:"url"`
	Referer string `json:"referer"`
	RequestURI string `json:"request_uri"`
	Profile ProfileData `json:"profile"`
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