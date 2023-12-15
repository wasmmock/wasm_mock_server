package util

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/google/martian"
)

type TraceRequest struct {
	Host            string `json:"host"`
	Path            string `json:"path"`
	Method          string `json:"method"`
	Minor           string `json:"minor"`
	Body            string `json:"body"`
	ContentLength   int64  `json:"content_length"`
	ContentEncoding string `json:"content_encoding"`
}
type TraceResponse struct {
	Minor           string `json:"minor"`
	Body            string `json:"body"`
	ContentLength   int64  `json:"content_length"`
	Time            int64  `json:"time"`
	Status          string `json:"status"`
	ContentEncoding string `json:"content_encoding"`
	ContentType     string `json:"content_type"`
}
type Trace struct {
	Request  TraceRequest  `json:"request"`
	Response TraceResponse `json:"response"`
}

func ConvertToTrace(resp *http.Response, readable_req_body string, readable_res_body string) Trace {
	req := resp.Request
	var minor_req = ""
	var minor_res = ""
	var duration = time.Second * 0
	gctx := martian.NewContext(resp.Request)

	if n, ok := gctx.Get("StartTime"); ok {
		startTime := n.(time.Time)
		duration = time.Now().Sub(startTime)
	}
	b, err := httputil.DumpRequestOut(req, false)
	if err == nil {
		minor_req = string(b)
	}
	b, err = httputil.DumpResponse(resp, false)
	if err == nil {
		minor_res = string(b)
	}

	if req.ContentLength > 0 && readable_req_body == "" {
		readable_req_body, err = HttpContentEncodingDecode(req.Header, req.Body, req.ContentLength, "req")
		if err != nil {
			readable_req_body = err.Error()
		}
	}
	if resp.ContentLength > 0 && readable_res_body == "" {
		readable_res_body, err = HttpContentEncodingDecode(resp.Header, resp.Body, resp.ContentLength, "res")
		if err != nil {
			readable_res_body = err.Error()
		}
	}
	return Trace{
		Request: TraceRequest{
			Host:            req.Host,
			Path:            req.URL.Path,
			Method:          req.Method,
			Body:            readable_req_body,
			Minor:           minor_req,
			ContentLength:   req.ContentLength,
			ContentEncoding: req.Header.Get("Content-Encoding"),
		},
		Response: TraceResponse{
			Minor:           minor_res,
			Body:            readable_res_body,
			Status:          resp.Status,
			Time:            duration.Milliseconds(),
			ContentLength:   resp.ContentLength,
			ContentEncoding: resp.Header.Get("Content-Encoding"),
			ContentType:     resp.Header.Get("Content-Type"),
		},
	}
}
