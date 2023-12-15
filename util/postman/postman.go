package postman

import (
	"net/http"
)

type Info struct {
	Postman_id string `json:"_postman_id"`
	Name       string `json:"name"`
	Schema     string `json:"schema"`
}
type Query struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type Url struct {
	Raw      string   `json:"raw"`
	Protocol string   `json:"protocol"`
	Host     []string `json:"host"`
	Path     []string `json:"path"`
	Query    []Query  `json:"query"`
}

type Request struct {
	Method string      `json:"method"`
	Header http.Header `json:"header"`
	Url    Url         `json:"url"`
}
type Item struct {
	Name     string     `json:"name"`
	Request  Request    `json:"request"`
	Response []Response `json:"response"`
}
type Response struct {
	Name            string   `json:"name"`
	OriginalRequest string   `json:"originalRequest"`
	Status          string   `json:"status"`
	Header          string   `json:"header"`
	Cookie          []Cookie `json:"cookie"`
	Body            string   `json:"body"`
}
type Cookie struct {
	Domain         string `json:"domain"`
	Expires        int32  `json:"expires"`
	HostOnly       bool   `json:"hostOnly"`
	HttpOnly       bool   `json:"httpOnly"`
	Key            string `json:"key"`
	Path           string `json:"path"`
	Secure         bool   `json:"secure"`
	Session        bool   `json:"session"`
	PostmanStoreId string `json:"_postman_storeId"`
	Value          string `json:"value"`
}
type Collection struct {
	Info Info   `json:"info"`
	Item []Item `json:"item"`
}

func ConvertItem(req *http.Request, name string) Item {
	var q = []Query{}
	for k, v := range req.URL.Query() {
		q = append(q, Query{Key: k, Value: v[0]})
	}
	return Item{
		Name: name,
		Request: Request{
			Method: req.Method,
			Header: req.Header,
			Url: Url{
				Raw:      req.URL.RawPath,
				Protocol: req.URL.Scheme,
				Host:     []string{req.URL.Host},
				Path:     []string{req.URL.Path},
				Query:    q,
			},
		},
		Response: []Response{},
	}
}
