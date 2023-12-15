package capabilities

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/wasmmock/wasm_mock_server/util/postman"

	"github.com/wasmmock/wasm_mock_server/util"
	"github.com/andybalholm/brotli"
)

type Response struct {
	HttpHeader  http.Header            `json:"http_header"`
	HttpCookie  map[string]string      `json:"http_cookie"`
	HttpBody    map[string]interface{} `json:"http_body"`
	HttpBodyRaw string                 `json:"http_body_raw"`
	Error       string                 `json:"error"`
	StatusCode  string                 `json:"status_code"`
	HttpRequest RequestReceivedInMock  `json:"http_req"`
}
type Request struct {
	Http1x   string `json:"http1x"`
	ProxyUrl string `json:"proxy_url"`
	HttpBody []byte `json:"http_body"`
}
type RequestForReport struct {
	Http1x   string `json:"http1x"`
	ProxyUrl string `json:"proxy_url"`
	HttpBody string `json:"http_body"`
}
type RequestReceivedInMock struct {
	HttpParam    map[string][]string    `json:"http_param"`
	HttpHeader   http.Header            `json:"http_header"`
	HttpCookie   map[string]string      `json:"http_cookie"`
	HttpBody     map[string]interface{} `json:"http_body"`
	HttpBodyRaw  string                 `json:"http_body_raw"`
	HttpProxyUrl string                 `json:"http_proxy_url"`
	HttpPath     string                 `json:"http_path"`
	HttpScheme   string                 `json:"http_scheme"`
	HttpMethod   string                 `json:"http_method"`
}
type FiddlerAB struct {
	ResA    Response `json:"res_a"`
	ResB    Response `json:"res_b"`
	UrlPath string   `json:"url_path"`
}

func HttpRequest(addr string, req []byte) ([]byte, util.CurlCommand, postman.Item, error) {
	wasmReq := Request{}
	err := json.Unmarshal(req, &wasmReq)
	data := bytes.NewReader([]byte(wasmReq.Http1x))
	r := bufio.NewReader(data)
	var response = make([]byte, 0)
	httpReq, err := http.ReadRequest(r)
	var curlCommand = util.CurlCommand{}
	var postmanItem = postman.Item{}
	if err != nil {
		return response, curlCommand, postmanItem, err
	}
	u, err := url.Parse(addr)
	if err != nil {
		return response, curlCommand, postmanItem, err
	}
	httpReq.RequestURI = ""
	httpReq.URL = u
	snapshot := bytes.NewBuffer(wasmReq.HttpBody)
	rc := ioutil.NopCloser(snapshot)
	httpReq.Body = rc
	httpReq.ContentLength = int64(snapshot.Len())
	httpReq.GetBody = func() (io.ReadCloser, error) {
		r := snapshot
		return ioutil.NopCloser(r), nil
	}
	if err != nil {
		return response, curlCommand, postmanItem, err
	}
	postmanItem = postman.ConvertItem(httpReq, "")
	if cu, er := util.GetCurlCommand(httpReq, wasmReq.HttpBody); er == nil {
		log.Println(cu.String())
		curlCommand = *cu
	} else {
		log.Println(er, httpReq.URL)
	}
	var client = &http.Client{}
	if wasmReq.ProxyUrl != "" {
		proxyUrl, err := url.Parse(wasmReq.ProxyUrl)
		if err != nil {
			return response, curlCommand, postmanItem, err
		}
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return response, curlCommand, postmanItem, err
	}
	cookies := map[string]string{}
	for _, c := range resp.Cookies() {
		cookies[c.Name] = c.Value
	}
	defer resp.Body.Close()
	data2, _ := ioutil.ReadAll(resp.Body)
	retRes := Response{
		HttpHeader: resp.Header,
		HttpCookie: cookies,
	}
	err = json.Unmarshal(data2, &retRes.HttpBody)
	if err != nil {
		retRes.Error = err.Error()
		retRes.HttpBodyRaw = string(data2)
	}
	retRes.StatusCode = resp.Status
	// client.Jar.Cookies()
	// client.
	req_cookies := map[string]string{}
	for _, c := range httpReq.Cookies() {
		req_cookies[c.Name] = c.Value
	}

	retRes.HttpRequest = RequestReceivedInMock{
		HttpParam:    httpReq.URL.Query(),
		HttpHeader:   httpReq.Header,
		HttpCookie:   req_cookies,
		HttpPath:     u.String(),
		HttpBody:     make(map[string]interface{}),
		HttpBodyRaw:  "",
		HttpProxyUrl: wasmReq.ProxyUrl,
		HttpMethod:   httpReq.Method,
	}
	retResBytes, err := json.Marshal(retRes)
	return retResBytes, curlCommand, postmanItem, err
}
func HttpRequestRaw(req http.Request) (Response, RequestReceivedInMock, error) {
	var client = &http.Client{}
	req.RequestURI = ""
	httpRes := &Response{}
	httpReqA := &RequestReceivedInMock{}
	respA, err := client.Do(&req)
	if err != nil {
		fmt.Print("err", err.Error())
		return *httpRes, *httpReqA, err
	}
	gzipBody, err := ioutil.ReadAll(respA.Body)
	if err != nil {
		fmt.Print("err", err.Error())
		return *httpRes, *httpReqA, err
	}
	var reader io.ReadCloser
	if strings.Contains(respA.Header.Get("Content-Encoding"), "gzip") {
		reader, err = gzip.NewReader(bytes.NewBuffer(gzipBody))
		defer reader.Close()
	} else if strings.Contains(respA.Header.Get("Content-Encoding"), "br") {
		var b = make([]byte, 60000)
		n, nerr := brotli.NewReader(bytes.NewBuffer(gzipBody)).Read(b)
		if nerr != nil {
			fmt.Println("brotli err %s", nerr.Error())
		}
		if n < 60000 {
			data := b[:n]
			reader = io.NopCloser(bytes.NewReader(data))
		} else {
			fmt.Println("n > 60000")
			return *httpRes, *httpReqA, err
		}

	} else {
		reader = ioutil.NopCloser(bytes.NewReader(gzipBody))
	}
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return *httpRes, *httpReqA, err
	}
	var httpBody map[string]interface{} = make(map[string]interface{})
	cookies := map[string]string{}
	for _, c := range respA.Cookies() {
		cookies[c.Name] = c.Value
	}

	if err := json.Unmarshal(body, &httpBody); err == nil {
		httpRes = &Response{
			HttpCookie:  cookies,
			HttpHeader:  respA.Header,
			HttpBody:    httpBody,
			HttpBodyRaw: "",
			Error:       "",
			StatusCode:  respA.Status,
		}
	} else {
		httpRes = &Response{
			HttpCookie:  cookies,
			HttpHeader:  respA.Header,
			HttpBody:    make(map[string]interface{}),
			HttpBodyRaw: string(body),
			Error:       err.Error(),
			StatusCode:  respA.Status,
		}
	}

	var httpBodyReq map[string]interface{} = make(map[string]interface{})
	cookiesReq := map[string]string{}
	for _, c := range req.Cookies() {
		cookiesReq[c.Name] = c.Value
	}
	if err != nil {

	}
	gzipReqBody, err := ioutil.ReadAll(req.Body)
	var reqReader io.ReadCloser
	switch req.Header.Get("Content-Encoding") {
	case "gzip":
		reqReader, err = gzip.NewReader(bytes.NewBuffer(gzipReqBody))
		defer reqReader.Close()
	default:
		reqReader = ioutil.NopCloser(bytes.NewReader(gzipBody))
	}
	reqBody, err := ioutil.ReadAll(reqReader)
	httpPath := req.URL.Path
	var http_param = make(map[string][]string)
	for k, v := range req.URL.Query() {
		http_param[k] = v
	}
	if err := json.Unmarshal(reqBody, &httpBodyReq); err == nil {

		httpReqA = &RequestReceivedInMock{
			HttpParam:    http_param,
			HttpCookie:   cookiesReq,
			HttpBody:     httpBodyReq,
			HttpBodyRaw:  "",
			HttpProxyUrl: "",
			HttpHeader:   req.Header,

			HttpPath:   httpPath,
			HttpMethod: req.Method,
			HttpScheme: req.URL.Scheme,
		}
	} else {
		httpReqA = &RequestReceivedInMock{
			HttpParam:    http_param,
			HttpCookie:   cookiesReq,
			HttpBody:     make(map[string]interface{}),
			HttpBodyRaw:  string(reqBody),
			HttpProxyUrl: "",
			HttpHeader:   req.Header,
			HttpPath:     httpPath,
			HttpMethod:   req.Method,
			HttpScheme:   req.URL.Scheme,
		}
	}
	return *httpRes, *httpReqA, err
}
