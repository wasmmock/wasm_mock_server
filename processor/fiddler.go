package processor

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wasmmock/wasm_mock_server/capabilities"
	"github.com/wasmmock/wasm_mock_server/util"
	"github.com/andybalholm/brotli"
	proto "github.com/golang/protobuf/proto"
	"github.com/google/martian"
)

type Modifier struct {
}

type fiddleResponseWriter struct {
	Response http.Response
}

func (c *fiddleResponseWriter) Header() http.Header {
	return c.Response.Header
}
func (c *fiddleResponseWriter) Write(payload []byte) (int, error) {
	return c.Response.Body.Read(payload)
}
func (c *fiddleResponseWriter) WriteHeader(statusCode int) {
	c.Response.StatusCode = statusCode
}

type buffer struct {
	bytes.Buffer
}

func (b *buffer) Read(payload []byte) (int, error) {
	return b.Buffer.Read(payload)
} // Add a Close method to our buffer so that we satisfy io.ReadWriteCloser.
func (b *buffer) Close() error {
	b.Buffer.Reset()
	return nil
}

type FiddleAB struct {
	ReqA        *http.Request
	ResB        capabilities.Response
	ReportIndex int64
	End         bool
}

func (e *Modifier) ModifyRequest(req *http.Request) error {
	//func(rw http.ResponseWriter, req *http.Request)
	if req.URL.Path == "/authority.cer" {
		return nil
	}
	if strings.ToLower(req.Method) == "options" {
		return nil
	}
	if len(req.URL.Path) > 0 {
		fmt.Println("req.URL.Path", req.URL.Path)
	}
	for key, value := range req.Header {
		if key == "Proxy-Remote-Addr" {
			req.RemoteAddr = value[0]
		} else if key == "Proxy-Host" {
			req.URL.Host = value[0]
			req.Host = value[0]
			req.URL.Scheme = "https"
		}
	}
	fiddlerBeforeRequestMapLongestFirst := []string{}
	fiddlerBeforeRequestMap.Range(func(key string, value string) bool {
		fiddlerBeforeRequestMapLongestFirst = append(fiddlerBeforeRequestMapLongestFirst, key)
		return true
	})
	// for k := range fiddlerBeforeRequestMap {
	// 	fiddlerBeforeRequestMapLongestFirst = append(fiddlerBeforeRequestMapLongestFirst, k)
	// }
	sort.Slice(fiddlerBeforeRequestMapLongestFirst, func(i, j int) bool {
		return len(fiddlerBeforeRequestMapLongestFirst[i]) > len(fiddlerBeforeRequestMapLongestFirst[j])
	})
	for _, k := range fiddlerBeforeRequestMapLongestFirst {
		var matches = false
		if strings.Contains(k, `\`) {
			var validID = regexp.MustCompile(k)
			if validID.MatchString(req.URL.Path) {
				matches = true
			}
		} else {
			matches = strings.Contains(req.URL.Path, k)
		}
		if matches && len(req.URL.Path) > 0 {
			httpReq := capabilities.RequestReceivedInMock{
				//HttpProxyUrl: req.URL.Scheme + "://" + req.URL.Host,
				HttpProxyUrl: req.URL.Host,
			}
			body := []byte{}

			if req.Body != nil {
				bodyn, err := ioutil.ReadAll(req.Body)
				body = bodyn
				req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyn))
				if err != nil {
					return err
				}
				var httpBody map[string]interface{} = make(map[string]interface{})
				if err := json.Unmarshal(body, &httpBody); err == nil {
					httpReq.HttpBody = httpBody
				} else {
					httpReq.HttpBodyRaw = string(body)
				}
			}

			cookies := map[string]string{}
			for _, c := range req.Cookies() {
				cookies[c.Name] = c.Value
			}
			httpReq.HttpMethod = req.Method
			httpReq.HttpPath = req.URL.Path
			httpReq.HttpScheme = req.URL.Scheme
			httpReq.HttpHeader = make(http.Header, len(req.Header))
			httpReq.HttpParam = make(map[string][]string, len(req.URL.Query()))
			httpReq.HttpCookie = cookies
			for key, value := range req.Header {
				if key != "Cookie" {
					httpReq.HttpHeader[key] = value
				}
			}
			for key, value := range req.URL.Query() {
				httpReq.HttpParam[key] = value
			}
			request, _ := json.Marshal(httpReq)
			responseBBytes := Bytes(request)
			response := &Bytes{}
			ctx := context.Background()
			//element_new := req.URL.Path
			request_index_in_report := int64(0)
			errcode, skipReport := FiddlerMockWasmBeforeReq(ctx, &responseBBytes, response, k, req.URL.Path, &request_index_in_report)
			_ = errcode
			req_b, er := proto.Marshal(response)
			req_c := make([]byte, len(req_b))
			copy(req_c, req_b)
			if er == nil {
				wasmReq := &capabilities.RequestReceivedInMock{}
				er1 := json.Unmarshal(req_c, wasmReq)
				if er1 != nil {
					fmt.Println("before_req err", er1)
					return er1
				}
				if er1 == nil {
					new_reqA_RawQuery := ""
					for k, v := range wasmReq.HttpParam {
						if new_reqA_RawQuery != "" {
							new_reqA_RawQuery = new_reqA_RawQuery + "&" + k + "=" + url.QueryEscape(v[0])
						} else {
							new_reqA_RawQuery = k + "=" + url.QueryEscape(v[0])
						}
					}
					for head, _ := range req.Header {
						req.Header.Del(head)
					}
					for head, val_Arr := range wasmReq.HttpHeader {
						for _, val := range val_Arr {

							req.Header.Add(head, val)
						}
					}

					for cookie_name, cookie_value := range wasmReq.HttpCookie {
						c := cookie_name + "=" + cookie_value
						expire := time.Now().AddDate(0, 0, 1)
						cookie := &http.Cookie{cookie_name, cookie_value, "/", req.Host, expire, expire.Format(time.UnixDate), 86400, true, true, 0, c, []string{c}}
						req.AddCookie(cookie)
					}
					if wasmReq.HttpProxyUrl != "" {
						req.URL.Host = wasmReq.HttpProxyUrl
						req.Host = wasmReq.HttpProxyUrl

					}

					if wasmReq.HttpScheme != "" {
						req.URL.Scheme = wasmReq.HttpScheme
					}
					///
					//req.RequestURI = req.URL.Path + "?" + new_reqA_RawQuery
					req.URL.RawQuery = new_reqA_RawQuery
					uID, ok := MockCommandUidMap.Get(k)
					if !skipReport && ok {
						command, err := util.GetCurlCommand(req, body)
						if err == nil {
							curlStep := util.Step{Pass: true, Description: command.String()}
							safeReports.SetStep(uID, curlStep, int(request_index_in_report))
						}
					}
					gctx := martian.NewContext(req)
					gctx.Set("ReqA", httpReq)
					gctx.Set("request_index_in_report", request_index_in_report)
					gctx.Set("skipReport", skipReport)
				}
			}
			break
		}
	}

	return nil
}

func (e *Modifier) ModifyResponse(resp *http.Response) error {
	if strings.ToLower(resp.Request.Method) == "options" {
		return nil
	}
	gctx := martian.NewContext(resp.Request)
	req := resp.Request
	if req.URL.Path == "/authority.cer" {
		ah := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: x509c.Raw,
		})
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(ah))

		resp.Header.Set("Content-Type", "application/x-x509-ca-cert")
		t := strconv.Itoa(len(ah))
		resp.Header.Set("Content-Length", t)
		resp.Header.Set("Content-Disposition", "inline; filename=mitmproxy-ca-cert.pem")
		return nil
	}
	if len(req.URL.Path) == 0 {
		return nil
	}
	var readable_resp_body = ""
	//var readable_req_body = ""
	//defer WsTrace(resp, &traceNetConn, readable_req_body, readable_resp_body)
	fiddlerBeforeRequestMapLongestFirst := []string{}

	fiddlerBeforeRequestMap.Range(func(key string, value string) bool {
		fiddlerBeforeRequestMapLongestFirst = append(fiddlerBeforeRequestMapLongestFirst, key)
		return true
	})
	sort.Slice(fiddlerBeforeRequestMapLongestFirst, func(i, j int) bool {
		return len(fiddlerBeforeRequestMapLongestFirst[i]) > len(fiddlerBeforeRequestMapLongestFirst[j])
	})
	for _, k := range fiddlerBeforeRequestMapLongestFirst {
		var matches = false
		if strings.Contains(k, `\`) {
			var validID = regexp.MustCompile(k)
			if validID.MatchString(req.URL.Path) {
				matches = true
			}
		} else {
			matches = strings.Contains(req.URL.Path, k)
		}
		if matches {
			if resp == nil {
				fmt.Println("ctx.Req.URL.Path resp is nil", req.URL.Path)
				return nil
			}
			//body := []byte{}
			defer resp.Body.Close()
			gzipBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("resp.Body err is not nil", err, resp.Body == nil)
				return nil
			}
			var reader io.ReadCloser
			if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
				reader, err = gzip.NewReader(bytes.NewBuffer(gzipBody))
			} else if strings.Contains(resp.Header.Get("Content-Encoding"), "br") {
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
					return nil
				}

			} else {
				reader = ioutil.NopCloser(bytes.NewReader(gzipBody))
			}
			if err != nil {
				fmt.Println("err____", err)
				return err
			}
			body, err := ioutil.ReadAll(reader)
			//resp.Body = ioutil.NopCloser(bytes.NewBuffer(gzipBody))
			if err != nil {
				fmt.Println("err____", err)
				return err
			}
			var httpBody map[string]interface{} = make(map[string]interface{})
			cookies := map[string]string{}
			for _, c := range resp.Cookies() {
				cookies[c.Name] = c.Value
			}
			httpRes := &capabilities.Response{}
			if err := json.Unmarshal(body, &httpBody); err == nil {
				httpRes = &capabilities.Response{
					HttpCookie:  cookies,
					HttpHeader:  make(http.Header, len(resp.Header)),
					HttpBody:    httpBody,
					HttpBodyRaw: "",
					Error:       "",
					StatusCode:  strconv.Itoa(resp.StatusCode),
					HttpRequest: capabilities.RequestReceivedInMock{},
				}

				readable_resp_body = string(body)
			} else {
				readable_resp_body = string(body)

				httpRes = &capabilities.Response{
					HttpCookie:  cookies,
					HttpHeader:  make(http.Header, len(resp.Header)),
					HttpBodyRaw: readable_resp_body,
					Error:       "",
					StatusCode:  strconv.Itoa(resp.StatusCode),
					HttpRequest: capabilities.RequestReceivedInMock{},
				}
			}
			httpRes.HttpRequest.HttpPath = resp.Request.URL.Path
			httpRes.HttpRequest.HttpScheme = resp.Request.URL.Scheme
			httpRes.HttpRequest.HttpHeader = make(http.Header, len(resp.Request.Header))
			httpRes.HttpRequest.HttpParam = make(map[string][]string, len(resp.Request.URL.Query()))
			httpRes.HttpRequest.HttpCookie = cookies
			for key, value := range resp.Request.Header {
				if key != "Cookie" {
					httpRes.HttpRequest.HttpHeader[key] = value
				}
			}
			for key, value := range resp.Request.URL.Query() {
				httpRes.HttpRequest.HttpParam[key] = value
			}
			for key, value := range resp.Header {
				httpRes.HttpHeader[key] = value
			}
			responseB, _ := json.Marshal(httpRes)
			responseBBytes := Bytes(responseB)
			response := &Bytes{}
			ctx := context.Background()
			if reqAinterface, ok := gctx.Get("ReqA"); ok {
				reqA := reqAinterface.(capabilities.RequestReceivedInMock)
				payload, err := json.Marshal(reqA.HttpBody)
				if err != nil {
					return err
				}
				var b bytes.Buffer
				if strings.Contains(req.Header.Get("Content-Encoding"), "gzip") {
					w := gzip.NewWriter(&b)
					w.Write(payload)
					w.Close()
				} else if strings.Contains(req.Header.Get("Content-Encoding"), "br") {
					w := brotli.NewWriter(&b)
					w.Write(payload)
					w.Close()
				} else {
					b.Write(payload)
				}
				body, err := ioutil.ReadAll(&b)
				data := bytes.NewReader(body)
				p := url.PathEscape(req.RequestURI)
				p1 := strings.ReplaceAll(p, "%2F", "/")
				new_reqA, er := http.NewRequest(reqA.HttpMethod, reqA.HttpScheme+"://"+reqA.HttpProxyUrl+p1, data)
				if er != nil {
					log.Println("http create New Req err panic", er, reqA.HttpMethod, "host", reqA.HttpProxyUrl, "rURI", p1)
					return er
				}
				host := req.URL.Host
				new_reqA.URL.Host = host
				ra := req.RemoteAddr
				new_reqA.RemoteAddr = ra
				for head, val_Arr := range reqA.HttpHeader {
					for _, val := range val_Arr {
						new_reqA.Header.Add(head, val)
					}
				}
				new_reqA_RawQuery := ""
				for k, v := range reqA.HttpParam {
					if new_reqA_RawQuery != "" {
						new_reqA_RawQuery = new_reqA_RawQuery + "&" + k + "=" + v[0]
					} else {
						new_reqA_RawQuery = k + "=" + v[0]
					}
				}
				new_reqA.RequestURI = reqA.HttpPath + "?" + new_reqA_RawQuery
				for cookie_name, cookie_value := range reqA.HttpCookie {
					c := cookie_name + "=" + cookie_value
					expire := time.Now().AddDate(0, 0, 1)
					cookie := &http.Cookie{cookie_name, cookie_value, "/", req.Host, expire, expire.Format(time.UnixDate), 86400, true, true, 0, c, []string{c}}
					new_reqA.AddCookie(cookie)
				}
				//zz
				if reqA.HttpProxyUrl != "" {
					new_reqA.URL.Host = reqA.HttpProxyUrl
				}
				if reqA.HttpScheme != "" {
					new_reqA.URL.Scheme = reqA.HttpScheme
				}
				request_index_in_report := int64(0)
				if n, ok := gctx.Get("request_index_in_report"); ok {
					request_index_in_report = n.(int64)
				}
				var skipReport = false
				if s, ok := gctx.Get("skipReport"); ok {
					skipReport = s.(bool)
				}
				FiddlerMockWasmBeforeRes(ctx, &responseBBytes, response, k, new_reqA, request_index_in_report, skipReport)

				httpRes = &capabilities.Response{}
				if bytes, err := proto.Marshal(response); err == nil {
					//crash
					bytes_c := make([]byte, len(bytes))
					copy(bytes_c, bytes)
					json.Unmarshal(bytes_c, httpRes)
				}
				for h, _ := range resp.Header {
					resp.Header.Del(h)
				}
				for h, v := range httpRes.HttpHeader {
					if len(v) > 0 {
						resp.Header.Add(h, v[0])
					}
				}
				for h, v := range httpRes.HttpCookie {
					resp.Header["Cookies"] = append(resp.Header["Cookies"], h+": "+v+";")
				}
				if httpRes.StatusCode != "" {
					if i, err := strconv.Atoi(httpRes.StatusCode); err == nil {
						resp.StatusCode = i
						resp.Status = http.StatusText(i)
					}
				}
				if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
					if httpRes.HttpBody != nil {
						if json_str, err := json.Marshal(httpRes.HttpBody); err == nil {
							var b bytes.Buffer
							w := gzip.NewWriter(&b)
							w.Write(json_str)
							w.Close()
							resp.Header.Set("Content-Length", strconv.Itoa(b.Len()))
							resp.ContentLength = int64(b.Len())
							fmt.Println("ContentLength", resp.ContentLength)
							resp.Body = ioutil.NopCloser(&b)
							return nil
						}
					}
				} else if strings.Contains(resp.Header.Get("Content-Encoding"), "br") {
					if httpRes.HttpBody != nil {
						if json_str, err := json.Marshal(httpRes.HttpBody); err == nil {
							var b bytes.Buffer
							w := brotli.NewWriter(&b)
							w.Write(json_str)
							w.Close()
							resp.Header.Set("Content-Length", strconv.Itoa(b.Len()))
							resp.ContentLength = int64(b.Len())
							fmt.Println("ContentLength", resp.ContentLength)
							resp.Body = ioutil.NopCloser(&b)
							return nil
						}
					}
				} else {
					if httpRes.HttpBody != nil {
						if json_str, err := json.Marshal(httpRes.HttpBody); err == nil {
							resp.Header.Set("Content-Length", strconv.Itoa(len(json_str)))
							resp.ContentLength = int64(len(json_str))
							resp.Body = ioutil.NopCloser(bytes.NewReader(json_str))
							return nil
						}
					} else if httpRes.HttpBodyRaw != "" {
						resp.Header.Set("Content-Length", strconv.Itoa(len(httpRes.HttpBodyRaw)))
						resp.ContentLength = int64(len(httpRes.HttpBodyRaw))
						resp.Body = ioutil.NopCloser(bytes.NewReader([]byte(httpRes.HttpBodyRaw)))
						return nil
					}
				}
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(gzipBody))

			}
			break
		}
	}
	return nil
}
