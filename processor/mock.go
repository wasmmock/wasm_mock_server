package processor

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/wasmmock/wasm_mock_server/capabilities"
	"github.com/wasmmock/wasm_mock_server/logger"
	"github.com/wasmmock/wasm_mock_server/model"
	Security "github.com/wasmmock/wasm_mock_server/security"
	"github.com/wasmmock/wasm_mock_server/tcpproxy"
	"github.com/wasmmock/wasm_mock_server/util"
	"github.com/audiolion/ipip"
	proto "github.com/golang/protobuf/proto"
	"github.com/google/martian/cors"
	"github.com/google/martian/mitm"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	wapc "github.com/wapc/wapc-go"
	"github.com/wapc/wapc-go/engines/wazero"
)

func SetMock(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)
	if !Security.X_Api_Key(req) {
		response.Header.Message = "Not Authorised"
		util.JsonResponse(rw, response)
		return
	}
	rw.Header().Set("Content-Type", "text/html; charset=ascii")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")

	response.Header.Message = "ok"
	code, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	targets := req.URL.Query()["targets"][0]
	targetList := strings.Split(targets, ",")
	uID := uuid.New().String()
	for _, target := range targetList {
		MockCommandMockUidMap.Store(target, uID)
	}
	wasmLangArr := req.URL.Query()["wasm_lang"]
	var wasmLang string = "rust"
	if len(wasmLangArr) > 0 {
		wasmLang = wasmLangArr[0]
	}
	ctx := context.Background()
	engine := wazero.Engine()
	module, err := engine.New(ctx, hostCall, code, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	// module, err := wapc.New(code, hostCall)
	//defer module.Close(ctx)
	var instance wapc.Instance
	if wasmLang == "go" {
		//instance, err = module.Instantiate("wasi_snapshot_preview1")
		instance, err = module.Instantiate(ctx)
	} else {
		//instance, err = module.Instantiate("wasi_unstable")
		instance, err = module.Instantiate(ctx)
	}
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	//cleanup mockUidInstanceMap
	for k, v := range mockUidInstanceMap.GetAll() {
		mockUidExist := false
		MockCommandMockUidMap.Range(func(key string, value string) bool {
			if key == k {
				mockUidExist = true
				return false
			}
			return true
		})
		if !mockUidExist {
			v.Close()
			mockUidInstanceMap.Delete(k)
		}
	}
	instanceref := util.NewSafeInstance(instance)
	mockUidInstanceMap.Set(uID, &instanceref)
	util.JsonResponse(rw, response)
	return
}

func SetMockTCP(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)
	response.Header.Message = "ok"
	code, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	targets := req.URL.Query()["targets"][0]
	targetList := strings.Split(targets, ",")
	wasmLangArr := req.URL.Query()["wasm_lang"]
	var wasmLang string = "rust"
	if len(wasmLangArr) > 0 {
		wasmLang = wasmLangArr[0]
	}
	ports := req.URL.Query()["ports"][0]
	portList := strings.Split(ports, ",")
	for i, port := range portList {
		ln, err := net.Listen("tcp", ":"+port)
		if err != nil {
			response.Header.Message = err.Error()
			util.JsonResponse(rw, response)
			return
		}
		element_new := targetList[i]
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					// handle error
					fmt.Println("tcp", err.Error())
				}
				source := ""
				traceid := ""
				payload := []byte{}
				response := &Bytes{}
				req_marshalling := &Bytes{}
				res_marshalling := &Bytes{}
				_, err = conn.Read(payload)
				request := Bytes(payload)
				if err == nil {
					ctx := context.Background()
					//errcode := CommonMockWasm(ctx, &request, response, req_marshalling, res_marshalling, element_new, source, traceid)

					errcode := CommonMockWasmV2(ctx, "tcp", &request, response, req_marshalling, res_marshalling, element_new, source, traceid)
					if errcode != 0 {
						fmt.Println("TCP errorcode", errcode)
					}
				}
			}
		}()
	}
	uID := uuid.New().String()
	for _, target := range targetList {
		MockCommandMockUidMap.Store(target, uID)
	}
	ctx := context.Background()
	engine := wazero.Engine()
	module, err := engine.New(ctx, hostCall, code, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	//module, err := wapc.New(code, hostCall)
	//defer module.Close(ctx)
	var instance wapc.Instance
	if wasmLang == "go" {
		//instance, err = module.Instantiate("wasi_snapshot_preview1")
		instance, err = module.Instantiate(ctx)
	} else {
		//instance, err = module.Instantiate("wasi_unstable")
		instance, err = module.Instantiate(ctx)
	}
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	//cleanup mockUidInstanceMap
	for k, v := range mockUidInstanceMap.GetAll() {
		mockUidExist := false
		MockCommandMockUidMap.Range(func(key string, value string) bool {
			if key == k {
				mockUidExist = true
				return false
			}
			return true
		})
		if !mockUidExist {
			v.Close()
			mockUidInstanceMap.Delete(k)
		}
	}
	instanceref := util.NewSafeInstance(instance)
	mockUidInstanceMap.Set(uID, &instanceref)
	util.JsonResponse(rw, response)
	return
}

func MockHTTPInit() {
	go func() {
		var httpMockRouter *mux.Router = mux.NewRouter()
		httpMockRouter.HandleFunc("/", CallMockHTTP(""))
		httpMockServer.Handler = httpMockRouter
		if err := httpMockServer.ListenAndServe(); err == nil {
			fmt.Println("httpMockServer start")
			logger.Infof("httpMockServer start")
		}
	}()
}
func CallMockHTTP(mock_http_type string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		element_new := req.URL.EscapedPath()
		source := ""
		traceid := ""
		response := &Bytes{}
		body, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return
		}
		var httpBody map[string]interface{} = make(map[string]interface{})
		cookies := map[string]string{}
		for _, c := range req.Cookies() {
			cookies[c.Name] = c.Value
		}
		httpReq := &capabilities.RequestReceivedInMock{}
		if err := json.Unmarshal(body, &httpBody); err == nil {
			httpReq = &capabilities.RequestReceivedInMock{
				HttpParam:  req.URL.Query(),
				HttpCookie: cookies,
				HttpHeader: req.Header,
				HttpBody:   httpBody,
				HttpPath:   req.URL.Path,
				HttpMethod: req.Method,
			}
		} else {
			httpReq = &capabilities.RequestReceivedInMock{
				HttpParam:    req.URL.Query(),
				HttpCookie:   cookies,
				HttpHeader:   req.Header,
				HttpBodyRaw:  string(body),
				HttpProxyUrl: "",
				HttpPath:     req.URL.Path,
				HttpMethod:   req.Method,
			}
		}

		request, err := json.Marshal(httpReq)
		requestBytes := Bytes(request)
		req_marshalling := &Bytes{}
		res_marshalling := &Bytes{}
		if err == nil {
			ctx := context.Background()
			errcode := uint32(0)
			//			errcode = CommonMockWasm(ctx, &requestBytes, response, req_marshalling, res_marshalling, element_new, source, traceid)
			errcode = CommonMockWasmV2(ctx, "http", &requestBytes, response, req_marshalling, res_marshalling, element_new, source, traceid)
			if errcode == 0 {
				b, er := proto.Marshal(response)
				if er == nil {
					res := &capabilities.Response{}
					er1 := json.Unmarshal(b, res)
					if er1 == nil {
						util.JsonResponse(rw, res.HttpBody)
						for key, element := range res.HttpHeader {
							rw.Header().Set(key, element[0])
						}
						for key, element := range res.HttpCookie {
							expire := time.Now().AddDate(0, 0, 1)
							c := key + "=" + element
							cookie := &http.Cookie{key, element, "/", "www.domain.com", expire, expire.Format(time.UnixDate), 86400, true, true, 0, c, []string{c}}
							http.SetCookie(rw, cookie)
						}
					}
				}
				req_b, er := proto.Marshal(req_marshalling)
				if er == nil {
					wasmReq := &capabilities.RequestReceivedInMock{}
					er1 := json.Unmarshal(req_b, wasmReq)
					if er1 == nil {
						zbody, err := json.Marshal(wasmReq.HttpBody)
						if err != nil {
							return
						}
						data := bytes.NewReader([]byte(zbody))
						new_req, err := http.NewRequest(req.Method, req.URL.Path, data)
						for head, val_Arr := range wasmReq.HttpHeader {
							for _, val := range val_Arr {
								new_req.Header.Add(head, val)
							}
						}
						for cookie_name, cookie_value := range wasmReq.HttpCookie {
							var found = false
							for _, existing_cookie := range req.Cookies() {
								if cookie_name == existing_cookie.Name {
									new_req.AddCookie(existing_cookie)
									found = true
									break
								}
							}
							if !found {
								c := cookie_name + "=" + cookie_value
								expire := time.Now().AddDate(0, 0, 1)
								cookie := &http.Cookie{cookie_name, cookie_value, "/", req.Host, expire, expire.Format(time.UnixDate), 86400, true, true, 0, c, []string{c}}
								new_req.AddCookie(cookie)
							}
						}
						if wasmReq.HttpProxyUrl != "" {
							new_req.Header.Add("Proxy-Url", wasmReq.HttpProxyUrl)
						}
						req = new_req
					}
				}
			} else {
				rw.Header().Set("HTTP-Code", fmt.Sprint(errcode))
				fmt.Println("Http errorcode", errcode)
			}
		}
		return
	}
}
func SetMockHTTP(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)
	response.Header.Message = "ok"
	code, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	targets := req.URL.Query()["targets"][0]
	targetList := strings.Split(targets, ",")
	wasmLangArr := req.URL.Query()["wasm_lang"]
	var wasmLang string = "rust"
	if len(wasmLangArr) > 0 {
		wasmLang = wasmLangArr[0]
	}
	uID := uuid.New().String()
	for _, target := range targetList {
		MockCommandMockUidMap.Store(target, uID)
		httpMockCommandMap[target] = uID
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	if err := httpMockServer.Shutdown(ctx); err != nil {
		fmt.Println("Server forced to shutdown:", err)
	}
	time.Sleep(1 * time.Second)
	var httpMockRouter *mux.Router = mux.NewRouter()
	for key, _ := range httpMockCommandMap {
		fmt.Println("key", key)
		httpMockRouter.HandleFunc(key, CallMockHTTP(""))
	}
	httpMockServerNew := http.Server{
		Addr:    httpMockServer.Addr,
		Handler: httpMockRouter,
	}
	httpMockServer = httpMockServerNew
	go func() {
		if err := httpMockServer.ListenAndServe(); err == nil {
			logger.Infof("httpMockServer start")
		}
	}()
	engine := wazero.Engine()
	module, err := engine.New(ctx, hostCall, code, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	//module, err := wapc.New(code, hostCall)
	defer module.Close(ctx)
	var instance wapc.Instance
	if wasmLang == "go" {
		//instance, err = module.Instantiate("wasi_snapshot_preview1")
		instance, err = module.Instantiate(ctx)
	} else {
		//instance, err = module.Instantiate("wasi_unstable")
		instance, err = module.Instantiate(ctx)
	}
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	//cleanup mockUidInstanceMap
	for k, v := range mockUidInstanceMap.GetAll() {
		mockUidExist := false
		MockCommandMockUidMap.Range(func(key string, value string) bool {
			if key == k {
				mockUidExist = true
				return false
			}
			return true
		})
		if !mockUidExist {
			v.Close()
			mockUidInstanceMap.Delete(k)
		}
	}
	instanceref := util.NewSafeInstance(instance)
	mockUidInstanceMap.Set(uID, &instanceref)
	util.JsonResponse(rw, response)
	return
}

// func configure(pattern string, handler http.Handler, mux *http.ServeMux) {
// 	allowCORS := false
// 	if allowCORS {
// 		handler = cors.NewHandler(handler)
// 	}

// 	// register handler for martian.proxy to be forwarded to
// 	// local API server
// 	mux.Handle(path.Join("martian.proxy", pattern), handler)

// }
// configure installs a configuration handler at path.
func configure(pattern string, handler http.Handler, mux *http.ServeMux) {
	allowCORS := false
	if allowCORS {
		handler = cors.NewHandler(handler)
	}

	// register handler for martian.proxy to be forwarded to
	// local API server
	mux.Handle(path.Join("martian.proxy", pattern), handler)

	// register handler for local API server
	p := path.Join("localhost:8181", pattern)
	mux.Handle(p, handler)
}
func GetCert(rw http.ResponseWriter, req *http.Request) {
	ah := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: x509c.Raw,
	})
	rw.Header().Set("Content-Disposition", "inline; filename=mitmproxy-ca-cert.pem")
	rw.Header().Set("Content-Type", "application/x-x509-ca-cert")
	rw.Write(ah)
	// t := strconv.Itoa(len(ah))
	// fmt.Println("content-length", t)
	// resp.Header.Set("Content-Length", t)
}
func MockFIDDLEInit() {
	addr := ":20810"
	port := os.Getenv("PORT_http_proxy")
	if port != "" {
		addr = ":" + port
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Start Mock Fiddler at " + addr)

	skipTLSVerify := false

	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLSVerify,
		},
	}
	fiddlerProxy.SetRoundTripper(tr)
	tlsc, err := tls.LoadX509KeyPair("cert_folder/proxy-ca.pem", "cert_folder/proxy-ca.key")
	if err != nil {
		log.Fatal(err)
	}

	var priv interface{}
	priv = tlsc.PrivateKey
	tls_port := os.Getenv("PORT_tls")
	tlsAddr := ":4043"
	if port != "" {
		tlsAddr = ":" + tls_port
	}
	validity := time.Hour
	organization := "Martian Proxy"
	x509c, err = x509.ParseCertificate(tlsc.Certificate[0])
	if err != nil {
		log.Fatal(err)
	}
	if x509c != nil && priv != nil {
		mc, err := mitm.NewConfig(x509c, priv)
		if err != nil {
			log.Fatal(err)
		}

		mc.SetValidity(validity)
		mc.SetOrganization(organization)
		mc.SkipTLSVerify(skipTLSVerify)

		fiddlerProxy.SetMITM(mc)
		// Start TLS listener for transparent MITM.
		tl, err := net.Listen("tcp", tlsAddr)
		if err != nil {
			log.Fatal(err)
		}

		go fiddlerProxy.Serve(tls.NewListener(tl, mc.TLS()))
	}
	m := &Modifier{}
	fiddlerProxy.SetRequestModifier(m)
	fiddlerProxy.SetResponseModifier(m)
	//m := martianhttp.NewModifier()
	// fg.AddRequestModifier(m)
	// fg.AddResponseModifier(m)
	go fiddlerProxy.Serve(l)
	log.Printf("martian: starting proxy on %s", l.Addr().String())

}
func SetMockBeforeRequestFiddler(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)
	response.Header.Message = "ok"
	code, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	targets := req.URL.Query()["targets"][0]
	targetList := strings.Split(targets, ",")
	wasmLangArr := req.URL.Query()["wasm_lang"]
	var wasmLang string = "rust"
	if len(wasmLangArr) > 0 {
		wasmLang = wasmLangArr[0]
	}
	uID := uuid.New().String()
	for _, target := range targetList {
		MockCommandMockUidMap.Store(target, uID)
		fiddlerBeforeRequestMap.Store(target, uID)
	}
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	// defer cancel()
	// fiddlerProxy.Close()
	// l, _ := net.Listen("tcp", "20810")
	// fiddlerProxy.Serve(l)
	ctx := context.Background()
	engine := wazero.Engine()
	module, err := engine.New(ctx, hostCall, code, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	//module, err := wapc.New(code, hostCall)
	//defer module.Close(ctx)
	var instance wapc.Instance
	if wasmLang == "go" {
		//instance, err = module.Instantiate("wasi_snapshot_preview1")
		instance, err = module.Instantiate(ctx)
	} else {
		//instance, err = module.Instantiate("wasi_unstable")
		instance, err = module.Instantiate(ctx)
	}
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	//cleanup mockUidInstanceMap
	for k, v := range mockUidInstanceMap.GetAll() {
		mockUidExist := false
		MockCommandMockUidMap.Range(func(key string, value string) bool {
			if key == k {
				mockUidExist = true
				return false
			}
			return true
		})
		if !mockUidExist {
			v.Close()
			mockUidInstanceMap.Delete(k)
		}
	}
	added_fn, err := instance.Invoke(ctx, "add_functions", []byte{})
	if err == nil {
		fmt.Println("added_fn", string(added_fn), err)
	} else {
		response.Header.Message = err.Error()
		fmt.Println("added_fn err", err.Error())
	}

	instanceref := util.NewSafeInstance(instance)
	mockUidInstanceMap.Set(uID, &instanceref)
	util.JsonResponse(rw, response)
	return
}
func SetMockBeforeRequestTCPFiddler(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)
	response.Header.Message = "ok"
	code, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	targets := req.URL.Query()["targets"][0]
	targetList := strings.Split(targets, ",")
	wasmLangArr := req.URL.Query()["wasm_lang"]
	var wasmLang string = "rust"
	if len(wasmLangArr) > 0 {
		wasmLang = wasmLangArr[0]
	}
	uID := uuid.New().String()
	ctx := context.Background()
	engine := wazero.Engine()
	module, err := engine.New(ctx, hostCall, code, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	//module, err := wapc.New(code, hostCall)
	if err != nil {
		log.Println("module err", err, "len", len(code))
	}
	for _, port_map := range targetList {
		port_map_arr := strings.Split(port_map, "-:")
		if len(port_map_arr) < 2 {
			response.Header.Message = "targets should be 3335-:3334,"
			util.JsonResponse(rw, response)
			return
		}
		localAddr := port_map_arr[0]
		if _, err := strconv.Atoi(localAddr); err == nil {
			localAddr = ":" + localAddr
		}
		laddr, err := net.ResolveTCPAddr("tcp", localAddr)
		if err != nil {
			response.Header.Message = err.Error()
			util.JsonResponse(rw, response)
			return
		}
		remoteAddr := port_map_arr[1]
		if _, err := strconv.Atoi(remoteAddr); err == nil {
			remoteAddr = ":" + remoteAddr
		}
		//unwrapTLS := false
		raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)

		if err != nil {
			response.Header.Message = err.Error()
			util.JsonResponse(rw, response)
			return
		}
		if v, ok := Security.SecurityMap["local_dial_range"]; ok && ipip.IsPrivate(raddr.IP) {
			r := strings.Split(v.ConfigValue, "-")
			if len(r) < 2 {
				return
			}
			r_0 := r[0]
			r_1 := r[1]
			if p_0, err := strconv.Atoi(r_0); err == nil {
				if p_1, err := strconv.Atoi(r_1); err == nil {
					if raddr.Port < p_0 || raddr.Port > p_1 {
						response.Header.Message = string(v.ErrorMsg)
						util.JsonResponse(rw, response)
						return
					}
				}
			}
		}
		MockCommandMockUidMap.Store(port_map, uID)
		fiddlerBeforeRequestMap.Store(port_map, uID)
		Before_req := func(payload []byte, port_map, l_remote_add string, real_l_remote_addr string, real_r_local_addr string) ([]byte, []capabilities.Entity) {
			request := Bytes(payload)
			_, e := TcpModifyRequest(&request, port_map, l_remote_add, real_l_remote_addr, real_r_local_addr)
			res, _ := proto.Marshal(&request)
			return res, e
		}
		Before_res := func(payload []byte, port_map, l_remote_add string, real_l_remote_addr string, real_r_local_addr string) ([]byte, []capabilities.Entity) {
			request := Bytes(payload)
			_, e := TcpModifyResponse(&request, port_map, l_remote_add, real_l_remote_addr, real_r_local_addr)
			res, _ := proto.Marshal(&request)
			return res, e
		}
		tcpproxy.PoolInit(uID, port_map, laddr, raddr, Before_req, Before_res, &MockCommandMockUidMap, &mockUidInstanceMap, module)
	}

	var instance wapc.Instance
	if wasmLang == "go" {
		//instance, err = module.Instantiate("wasi_snapshot_preview1")
		instance, err = module.Instantiate(ctx)
	} else {
		instance, err = module.Instantiate(ctx)
		//instance, err = module.Instantiate("wasi_unstable")
	}
	instance.Invoke(ctx, "add_functions", []byte{})
	instanceref := util.NewSafeInstance(instance)
	mockUidInstanceMap.Set(uID, &instanceref)
	util.JsonResponse(rw, response)
	return
}
