package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/wasmmock/wasm_mock_server/etc_creator"

	Admin "github.com/wasmmock/wasm_mock_server/admin"
	"github.com/wasmmock/wasm_mock_server/internal/box"
	"github.com/wasmmock/wasm_mock_server/logger"
	"github.com/wasmmock/wasm_mock_server/processor"
	Security "github.com/wasmmock/wasm_mock_server/security"
	"github.com/wasmmock/wasm_mock_server/tcpproxy"
	"github.com/google/martian"
	"github.com/gorilla/mux"
)

var tmpl template.Template

func NewServer(rpc_able processor.RpcAble, regCmd string) {

	martian.Init()
	k := strings.Split(regCmd, ",")
	processor.RegisterHostCall(rpc_able.HostCallSpecific)
	rpc_able.Start(k)
	processor.TemplateInit()
	processor.SetBaseWsMockFromBox()
	processor.SetBaseWsCallFromBox()
	etc_creator.CreateReportFolder()
	etc_creator.CreateCertfolder()
	handleHTTP(rpc_able)
}
func RecoverWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
func handleHTTP(rpc_able processor.RpcAble) {
	logger.Infof("Starting HTTP server...")
	addr := ":20825"
	port := os.Getenv("PORT_client")
	if port != "" {
		addr = ":" + port
	}

	rtr := mux.NewRouter()
	rtr.HandleFunc("/call/http", processor.CallHttp())
	rtr.HandleFunc("/call/rpc", processor.CallRpc(rpc_able))
	rtr.HandleFunc("/call/mysql", processor.CallMysql())
	rtr.HandleFunc("/call/raw_rpc", processor.CallRawRpc(rpc_able))
	rtr.HandleFunc("/call/fiddler", processor.CallFiddler(rpc_able))
	rtr.HandleFunc("/call/tcp_fiddler", processor.CallTcpFiddler(rpc_able))
	rtr.HandleFunc("/call/set_mock", processor.SetMock)
	rtr.HandleFunc("/call/set_mock_tcp", processor.SetMockTCP)
	rtr.HandleFunc("/call/set_mock_http", processor.SetMockHTTP)
	rtr.HandleFunc("/call/set_mock_fiddler", processor.SetMockBeforeRequestFiddler)
	rtr.HandleFunc("/call/set_mock_tcp_fiddler", processor.SetMockBeforeRequestTCPFiddler)
	rtr.HandleFunc("/call/v2/unified", processor.UnifiedV2(rpc_able))
	rtr.HandleFunc("/call/v2/call", processor.CallV2)
	rtr.HandleFunc("/call/set_cert", processor.SetCertFile)
	rtr.HandleFunc("/call/use_key", processor.UseKeyFile)
	rtr.HandleFunc("/call/set_base_ws_mock", processor.SetBaseWsMock)
	rtr.HandleFunc("/call/set_base_ws_call", processor.SetBaseWsCall)
	rtr.HandleFunc("/rpc_get", processor.CommandsGet())
	rtr.HandleFunc("/rpc_restart", processor.RegisterCommands(rpc_able))
	rtr.HandleFunc("/report/{uid}", processor.Report)
	rtr.HandleFunc("/report_data/{uid}", processor.ReportData)
	rtr.HandleFunc("/ws/trace", processor.WebsocketTrace)
	rtr.HandleFunc("/ws/rpc", processor.WebsocketRPC(rpc_able))
	rtr.HandleFunc("/ws/set_mock", processor.WebsocketSetMock)
	rtr.HandleFunc("/ws/set_mock_fiddler", processor.WebsocketSetMockBeforeRequestFiddler)
	rtr.HandleFunc("/ws/set_mock_tcp_fiddler", processor.WebsocketSetMockTcpFiddler)
	rtr.HandleFunc("/indexdb/store", processor.StoreIndexDB)
	rtr.HandleFunc("/indexdb/get", processor.GetIndexDB)
	rtr.HandleFunc("/cert/pem", processor.GetCert)
	rtr.HandleFunc("/admin/get_status", Admin.GetStatus())
	if v, ok := Security.SecurityMap["Remote-Verification"]; ok {
		var config = Security.RemoteVerificationConfig{}
		if err := json.Unmarshal([]byte(v.ConfigValue), &config); err != nil {
			fmt.Println("Security.SecurityMap[Remote-Verification] error: ")
		}
		if config.ServeEndpointLocal {
			rtr.HandleFunc("/login/2/authorization/google", processor.OauthLogin(rpc_able))
		}
	}
	rtr.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) { panic("test panic") })
	//rtr.PathPrefix("/").Handler(http.FileServer(http.Dir("public")))
	rtr.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		http.ServeFile(w, r, "public/index.html")
	})

	processor.MockHTTPInit()
	time.Sleep(2 * time.Second)
	processor.MockFIDDLEInit()
	tcpproxy.HandleRemove2()
	tcpproxy.SelfRemove()
	for filename, content := range box.Map() {
		filename2 := filename
		content2 := content
		rtr.HandleFunc(fmt.Sprintf("%v", filename),
			func(w http.ResponseWriter, r *http.Request) {
				http.ServeContent(w, r, "public"+filename2, time.Now(),
					bytes.NewReader(content2))
			})
	}

	logger.Infof("Http Server on: " + addr)
	//fs := http.FileServer(http.Dir("public"))
	// if err := http.ListenAndServe(addr, Security.X_Api_Key_Middleware(rtr)); err != nil {
	// 	logger.Errorf("Failed to listen HTTP | err=%s", err.Error())
	// }
	if err := http.ListenAndServe(addr, rtr); err != nil {
		logger.Errorf("Failed to listen HTTP | err=%s", err.Error())
	}
	return
}
func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("middleware", r.URL)
		h.ServeHTTP(w, r)
	})
}
func logHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := httptest.NewRecorder()
		fn(rec, r)
		logger.Infof("url:%v, method:%v, req:%v, resp:%v", r.URL, r.Method, r.Body, rec.Body)

		w.WriteHeader(rec.Code)
		_, err := rec.Body.WriteTo(w)
		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			return
		}
	}
}
