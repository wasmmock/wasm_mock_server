package processor

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wasmmock/wasm_mock_server/capabilities"
	Security "github.com/wasmmock/wasm_mock_server/security"

	"github.com/wasmmock/wasm_mock_server/logger"
	"github.com/wasmmock/wasm_mock_server/model"
	"github.com/wasmmock/wasm_mock_server/util"
	proto "github.com/golang/protobuf/proto"
	"github.com/google/martian"
	"github.com/google/uuid"
	"github.com/tebeka/selenium"
	"github.com/vmihailenco/msgpack"
	wapc "github.com/wapc/wapc-go"
	"github.com/wapc/wapc-go/engines/wazero"
)

type Bytes []byte

func (*Bytes) ProtoMessage() {}

func (m *Bytes) Marshal() (dAtA []byte, err error) {
	return *m, nil
}
func (m *Bytes) Unmarshal(dAtA []byte) error {
	*m = dAtA
	return nil
}
func (m *Bytes) Reset() { *m = Bytes{} }

func (m *Bytes) String() string {
	return proto.CompactTextString(m)
}
func (m *Bytes) Size() {}
func (m *Bytes) UnmarshalJSON(dAtA []byte) error {
	*m = dAtA
	return nil
}
func (m *Bytes) MarshalJSON() ([]byte, error) {
	return *m, nil
}

var indexMap map[string]int64 = make(map[string]int64)             //uid:int
var MockCommandUidMap util.SafeStringMap = util.NewSafeStringMap() //mock command:uid string
var safeReports = util.SafeReports{Reports: make(map[string]util.Report, 0)}
var mockUidInstanceMap util.SafeInstanceMap = util.NewSafeInstanceMap() //mock uid:Instance
var MockCommandMockUidMap util.SafeStringMap = util.NewSafeStringMap()  //mockCommand:Mock uid
var traceUidMap util.SafeStringMap = util.NewSafeStringMap()
var backgroundMock []util.Mock = []util.Mock{}
var hits util.SafeHitMap = util.NewSafeHitMap()
var hitcount int32 = 0
var RpcHandler RpcAble
var registeredCommands = []string{}
var wdMap map[string]selenium.WebDriver = make(map[string]selenium.WebDriver)
var wdServiceMap map[string]*selenium.Service = make(map[string]*selenium.Service)
var wdElementMap map[string]selenium.WebElement = make(map[string]selenium.WebElement)
var indexDBMap util.SafeIndexDb = util.NewSafeIndexDb()
var httpMockServer http.Server = http.Server{
	Addr: ":20822",
}

var httpMockCommandMap map[string]string = make(map[string]string)

type httpMockChanProtocol struct {
	action string
}

var httpMockChan chan httpMockChanProtocol = make(chan httpMockChanProtocol)

var fiddlerCertMap map[string]tls.Certificate = make(map[string]tls.Certificate)
var fiddlerBeforeRequestMap util.SafeStringMap = util.NewSafeStringMap()
var fiddlerProxy *martian.Proxy = martian.NewProxy()
var fiddlerQueue map[string]chan FiddleAB = make(map[string]chan FiddleAB)
var fiddlerQueueOut map[string]chan FiddleAB = make(map[string]chan FiddleAB)
var fiddlerUniqueUrlPath map[string][]string = make(map[string][]string)
var tcpFiddlerQueue map[string]chan TcpFiddleAB = make(map[string]chan TcpFiddleAB)
var tcpFiddlerQueueOut map[string]chan TcpFiddleAB = make(map[string]chan TcpFiddleAB)
var tcpNewWriteReqChan map[string]chan capabilities.TcpReq = make(map[string]chan capabilities.TcpReq)
var tcpNewWriteReqResponderChan map[string]chan capabilities.TcpReq = make(map[string]chan capabilities.TcpReq)

var tcpNewWriteResChan map[string]chan capabilities.TcpReq = make(map[string]chan capabilities.TcpReq)
var tcpRequestIdMap map[string]map[string]capabilities.TcpReq = make(map[string]map[string]capabilities.TcpReq)
var tcpLocalListener map[string]*net.TCPListener = make(map[string]*net.TCPListener)
var x509c *x509.Certificate

func saveHandleInitErr(rw http.ResponseWriter, response *model.CallApiResponse, what string, err error, uid string) error {
	if err != nil {
		if safeReports.Len(uid) == 0 {
			safeReports.AppendUnitTest(uid, 0)
		}
		response.Header = model.Header{
			Message:  what + " " + err.Error(),
			ReportId: uid,
		}
		util.JsonResponse(rw, response)
		iStep := util.Step{Pass: false, Description: what + " " + err.Error()}
		safeReports.AppendStep(uid, iStep)
		safeReports.AppendEnd(uid, backgroundMock, hits.Clone())
	}
	return err
}
func saveHandleLoopErr(rw http.ResponseWriter, response *model.CallApiResponse, what string, err error, uid string) error {
	if err != nil {
		iStep := util.Step{Pass: false, Description: what + " " + err.Error()}
		safeReports.AppendStep(uid, iStep)
		safeReports.AppendEnd(uid, backgroundMock, hits.Clone())
	}
	return err
}
func saveHandleMockErr(command string, reqJSON []byte, resJSON []byte, startTime time.Time, index int64, source string, traceid string, what string, err error, uid string) error {
	if err != nil {
		endTime := time.Now()
		duration := util.DurationToString(endTime.Sub(startTime))
		mo := util.Mock{
			Command:  command,
			Response: "command" + command + " req:" + string(resJSON) + " | what: " + what + " | " + err.Error(),
			Request:  string(reqJSON),
			Duration: duration,
			Index:    index,
			Pass:     false,
			EndTime:  endTime.Format("15:05:05"),
			Source:   source,
			TraceId:  traceid,
		}
		newBackgroundMock := []util.Mock{}
		for _, j := range backgroundMock {
			if startTime.Sub(j.Time).Milliseconds() < 5000 {
				newBackgroundMock = append(newBackgroundMock, j)
			}
		}
		safeReports.SetMock(uid, mo, int(index))
		newBackgroundMock = append(newBackgroundMock, mo)
		backgroundMock = newBackgroundMock
	}
	return err
}
func resetState(mockTargetsList []string, uID string) {
	fmt.Println("resetState---------", MockCommandUidMap)
	indexMap[uID] = 0
	for _, target := range mockTargetsList {
		indexMap[uID] = 0
		MockCommandUidMap.Delete(target)
	}
	delete(indexMap, uID)
	capabilities.SeleniumClose(&wdMap, &wdServiceMap, &wdElementMap, uID)
}
func BoundFromLoop(loop string, perviousError error) (int64, int64, error) {
	lower := int64(0)
	upper := int64(0)
	if strings.Contains(loop, ",") {
		loopArr := strings.Split(loop, ",")
		if len(loopArr) > 1 {
			lower, perviousError = strconv.ParseInt(loopArr[0], 10, 64)
			upper, perviousError = strconv.ParseInt(loopArr[1], 10, 64)
		}
	} else {
		upper, perviousError = strconv.ParseInt(loop, 10, 64)
	}
	return lower, upper, perviousError
}
func CallHttp() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// response := new(model.CallApiResponse)
		// if !Security.X_Api_Key(req) {
		// 	response.Header.Message = "Not Authorised"
		// 	util.JsonResponse(rw, response)
		// 	return
		// }
		res := model.CallApiResponse{
			Header: model.Header{},
		}
		code, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			util.JsonResponse(rw, res)
		}

		loop := "0"
		if len(req.URL.Query()["loop"]) > 0 {
			loop = req.URL.Query()["loop"][0]
		}
		mockTargets := ""
		if len(req.URL.Query()["targets"]) > 0 {
			mockTargets = req.URL.Query()["targets"][0]
		}
		wasmLangArr := req.URL.Query()["wasm_lang"]
		var wasmLang string = "rust"
		if len(wasmLangArr) > 0 {
			wasmLang = wasmLangArr[0]
		}
		wsWriter_ := httpWriterHTTP{
			loop_:         loop,
			rw_:           rw,
			response_:     res,
			mock_targets_: mockTargets,
			code:          code,
			wasm_lang_:    wasmLang,
		}
		CallWasm(&wsWriter_, "")
		return
	}
}

func CallRpc(rpcHandler RpcAble) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		res := model.CallApiResponse{}
		// if !Security.X_Api_Key(req) {
		// 	res.Header.Message = "Not Authorised"
		// 	util.JsonResponse(rw, res)
		// 	return
		// }
		code, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			util.JsonResponse(rw, res)
		}

		loop := "0"
		if len(req.URL.Query()["loop"]) > 0 {
			loop = req.URL.Query()["loop"][0]
		}
		mockTargets := ""
		if len(req.URL.Query()["mock_targets"]) > 0 {
			mockTargets = req.URL.Query()["mock_targets"][0]
		}
		wasmLangArr := req.URL.Query()["wasm_lang"]
		var wasmLang string = "rust"
		if len(wasmLangArr) > 0 {
			wasmLang = wasmLangArr[0]
		}
		wsWriter_ := httpWriterRPC{
			loop_:         loop,
			rw_:           rw,
			response_:     res,
			mock_targets_: mockTargets,
			wasm_lang_:    wasmLang,
			code:          code,
			rpc_able:      rpcHandler,
		}
		CallWasm(&wsWriter_, "")
		return
	}
}
func CallMysql() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		res := model.CallApiResponse{
			Header: model.Header{},
		}
		code, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			util.JsonResponse(rw, res)
		}

		loop := "0"
		if len(req.URL.Query()["loop"]) > 0 {
			loop = req.URL.Query()["loop"][0]
		}
		mockTargets := ""
		if len(req.URL.Query()["mock_targets"]) > 0 {
			mockTargets = req.URL.Query()["mock_targets"][0]
		}
		wsWriter_ := httpWriterHTTP{
			loop_:         loop,
			rw_:           rw,
			response_:     res,
			mock_targets_: mockTargets,
			code:          code,
		}
		CallWasm(&wsWriter_, "")
		return
	}
}
func CallRawRpc(rpcHandler RpcAble) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		//res := model.CallApiResponse{}
		// if !Security.X_Api_Key(req) {
		// 	res.Header.Message = "Not Authorised"
		// 	util.JsonResponse(rw, res)
		// 	return
		// }
		command := ""
		if len(req.URL.Query()["command"]) > 0 {
			command = req.URL.Query()["command"][0]
		}
		command = strings.ReplaceAll(command, "|", "?")
		code, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			rw.Header().Set("RPC-Code", "Readbody_err")
			return
		}
		rpcRes := Bytes{}
		rpcReq := Bytes(code)
		errCode := rpcHandler.RpcRequest(command, &rpcReq, &rpcRes)
		rw.Header().Set("RPC-Code", strconv.FormatUint(uint64(errCode), 10))
		rpcResBytes, err := rpcRes.Marshal()
		if err != nil {
			rw.Header().Set("RPC-Code", "ResUnmarshall_err")
			return
		}
		rw.Header().Set("Content-Type", "application/octet-stream")
		le, err2 := rw.Write(rpcResBytes)
		if err2 != nil {
			rw.Header().Set("RPC-Code", "ContentWrite_err")
		}
		logger.Infof("writelen %v %v", len(rpcResBytes), le)
	}
}
func CallFiddler(rpcHandler RpcAble) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		command := ""
		if len(req.URL.Query()["mock_targets"]) > 0 {
			command = req.URL.Query()["mock_targets"][0]
		}
		mockTargetsList := strings.Split(command, ",")
		duration := ""
		if len(req.URL.Query()["duration"]) > 0 {
			duration = req.URL.Query()["duration"][0]
		}
		duration_int, parseErr := strconv.ParseInt(duration, 10, 64)
		response := new(model.CallApiResponse)
		response.Header.Message = "added_fns: " + command
		if parseErr != nil {
			response.Header.Message = parseErr.Error()
			util.JsonResponse(rw, response)
		}
		uID := uuid.New().String()
		safeReports.ReportGen(uID)
		fiddlerQueue[uID] = make(chan FiddleAB)
		fiddlerQueueOut[uID] = make(chan FiddleAB)
		fiddlerUniqueUrlPath[uID] = make([]string, 0)
		go func() {
			rpcHandler.FiddleQueue(fiddlerQueue[uID], fiddlerQueueOut[uID])
		}()
		defer func() {
			delete(fiddlerQueue, uID)
			delete(fiddlerUniqueUrlPath, uID)
			//delete(fiddlerQueueOut, uID)
			safeReports.Save(uID)
		}()
		for _, target := range mockTargetsList {
			MockCommandUidMap.Store(target, uID)
		}
		safeReports.ReportGen(uID)
		duration_int_t := time.Duration(duration_int)
		time.Sleep(duration_int_t * time.Second)
		go func() {
			fmt.Println("1End....................")
			for _, target := range mockTargetsList {
				MockCommandUidMap.Delete(target)
			}
			// delete(fiddlerBeforeRequestMap,
			fiddlerQueue[uID] <- FiddleAB{End: true}
			fmt.Println("2End....................")
		}()

		code, _ := ioutil.ReadAll(req.Body)
		wasmLangArr := req.URL.Query()["wasm_lang"]
		var wasmLang string = "rust"
		if len(wasmLangArr) > 0 {
			wasmLang = wasmLangArr[0]
		}
		ctx := context.Background()
		engine := wazero.Engine()
		module, wasm_err := engine.New(ctx, hostCall, code, &wapc.ModuleConfig{
			Logger: wapc.PrintlnLogger,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		})
		//module, wasm_err := wapc.New(code, hostCall)
		var instance wapc.Instance
		if wasmLang == "go" {
			//instance, wasm_err = module.Instantiate("wasi_snapshot_preview1")
			instance, wasm_err = module.Instantiate(ctx)
		} else {
			//instance, wasm_err = module.Instantiate("wasi_unstable")
			instance, wasm_err = module.Instantiate(ctx)
		}
		if wasm_err != nil {
			instance.Invoke(ctx, "add_functions", []byte{})
			teardownInstance := util.NewSafeInstance(instance)
			ctx := context.Background()
			teardownInstance.Invoke(ctx, "teardown", []byte{})
			defer func() {
				teardownInstance.Invoke(ctx, "teardown", []byte{})
				module.Close(ctx)
			}()
		}
		for {
			v := <-fiddlerQueueOut[uID]
			if v.End {
				fmt.Println("breaking")
				break
			}
			reqA := v.ReqA
			resB := v.ResB
			index := v.ReportIndex
			resA, reqAA, err := capabilities.HttpRequestRaw(*reqA)
			reqAA_str, err := json.Marshal(reqAA)
			resA_str, err := json.Marshal(resA)
			startTime := time.Now()
			endTime := time.Now()
			duration := util.DurationToString(endTime.Sub(startTime))
			m := util.Mock{
				Command:  reqA.URL.Path,
				Request:  string(reqAA_str),
				Response: string(resA_str),
				Pass:     true,
				Duration: duration,
				EndTime:  endTime.Format("15:05:05"),
				Index:    index,
			}
			safeReports.SetMock(uID, m, int(index))
			if err == nil {
				fiddlerBeforeRequestMap.Range(func(k string, value string) bool {
					if strings.Contains(reqA.URL.Path, k) {
						if mock_uid, ok := MockCommandMockUidMap.Get(k); ok {
							if safe_instance, ok := mockUidInstanceMap.Get(mock_uid); ok {
								fiddlerAB := capabilities.FiddlerAB{
									ResA:    resA,
									ResB:    resB,
									UrlPath: reqA.URL.Path,
								}
								fiddlerAB_str, err := json.Marshal(fiddlerAB)
								if err == nil {
									ctx := context.WithValue(context.Background(), "report_index", int(index))
									ctx = context.WithValue(ctx, "uid", uID)
									_, err := safe_instance.Invoke(ctx, k+"_http_fiddler_ab", fiddlerAB_str)
									if err := saveHandleMockErr(reqA.URL.Path, []byte{}, []byte{}, startTime, index, "", "", "fiddler_ab ", err, uID); err != nil {
										fmt.Print("saveHandleMockErr ", err.Error(), string(fiddlerAB_str))
									}
								} else {
									fmt.Println("fiddlerAB marshal err", err)
								}
							}
						}
					}
					return true
				})

			}
		}

		response.Header.ReportId = uID
		util.JsonResponse(rw, response)
	}
}
func CallTcpFiddler(rpcHandler RpcAble) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		command := ""
		if len(req.URL.Query()["targets"]) > 0 {
			command = req.URL.Query()["targets"][0]
		}
		mockTargetsList := strings.Split(command, ",")
		duration := ""
		if len(req.URL.Query()["duration"]) > 0 {
			duration = req.URL.Query()["duration"][0]
		}
		duration_int, parseErr := strconv.ParseInt(duration, 10, 64)
		response := new(model.CallApiResponse)
		if parseErr != nil {
			response.Header.Message = parseErr.Error()
			util.JsonResponse(rw, response)
		}
		uID := uuid.New().String()
		safeReports.ReportGen(uID)
		tcpFiddlerQueue[uID] = make(chan TcpFiddleAB)
		tcpFiddlerQueueOut[uID] = make(chan TcpFiddleAB)
		fiddlerUniqueUrlPath[uID] = make([]string, 0)
		tcpRequestIdMap[uID] = make(map[string]capabilities.TcpReq)
		go func() {
			rpcHandler.TcpFiddleQueue(tcpFiddlerQueue[uID], tcpFiddlerQueueOut[uID])
		}()
		defer func() {
			delete(tcpFiddlerQueue, uID)
			delete(fiddlerUniqueUrlPath, uID)
			delete(tcpRequestIdMap, uID)
			//delete(fiddlerQueueOut, uID)
			safeReports.Save(uID)
		}()
		for _, target := range mockTargetsList {
			MockCommandUidMap.Store(target, uID)
		}
		safeReports.ReportGen(uID)
		duration_int_t := time.Duration(duration_int)
		time.Sleep(duration_int_t * time.Second)
		go func() {
			fmt.Println("1End....................")
			// for _, target := range mockTargetsList {
			// 	MockCommandUidMap.Delete(target)
			// }
			tcpFiddlerQueue[uID] <- TcpFiddleAB{End: true}
			fmt.Println("2End....................")
		}()

		code, _ := ioutil.ReadAll(req.Body)
		wasmLangArr := req.URL.Query()["wasm_lang"]
		var wasmLang string = "rust"
		if len(wasmLangArr) > 0 {
			wasmLang = wasmLangArr[0]
		}
		ctx := context.Background()
		engine := wazero.Engine()
		module, wasm_err := engine.New(ctx, hostCall, code, &wapc.ModuleConfig{
			Logger: wapc.PrintlnLogger,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		})
		//module, wasm_err := wapc.New(code, hostCall)
		var instance wapc.Instance
		if wasmLang == "go" {
			//instance, wasm_err = module.Instantiate("wasi_snapshot_preview1")
			instance, wasm_err = module.Instantiate(ctx)
		} else {
			//instance, wasm_err = module.Instantiate("wasi_unstable")
			instance, wasm_err = module.Instantiate(ctx)
		}
		if wasm_err != nil {
			teardownInstance := util.NewSafeInstance(instance)
			ctx := context.Background()
			teardownInstance.Invoke(ctx, "teardown", []byte{})
			defer func() {
				teardownInstance.Invoke(ctx, "teardown", []byte{})
				module.Close(ctx)
			}()
		}
		for {
			v := <-tcpFiddlerQueueOut[uID]
			if v.End {
				fmt.Println("breaking")
				break
			}
			//reqA := v.ReqA
			//resB := v.ResB
			index := v.ReportIndex
			startTime := time.Now()
			endTime := time.Now()
			duration := util.DurationToString(endTime.Sub(startTime))
			m := util.Mock{
				Command:  "",
				Request:  v.ReqAString,
				Response: v.ResBString,
				Pass:     true,
				Duration: duration,
				EndTime:  endTime.Format("15:05:05"),
				Index:    index,
			}
			safeReports.SetMock(uID, m, int(index))
			if true {
				fiddlerBeforeRequestMap.Range(func(k string, value string) bool {
					if mock_uid, ok := MockCommandMockUidMap.Get(k); ok {
						if safe_instance, ok := mockUidInstanceMap.Get(mock_uid); ok {
							ctx := context.Background()
							sDec := b64.StdEncoding.EncodeToString(v.ReqA)
							tcp_req := capabilities.TcpReq{
								Payload:    sDec,
								ReportType: "mock",
							}
							payload, err := msgpack.Marshal(tcp_req)
							if err != nil {
								fmt.Println("tcp_req err", err)
								return true
							}
							tcp_res_bytes, err := hostCall(ctx, uID, "foo", "tcp_new_req", payload)
							if err == nil {
								var tcp_res = capabilities.TcpReq{}
								if err := msgpack.Unmarshal(tcp_res_bytes, &tcp_res); err == nil {
									sDec, _ := b64.StdEncoding.DecodeString(tcp_res.Payload)
									fiddlerAB := capabilities.TcpFiddlerAB{
										ResA: sDec,
										ResB: v.ResB,
									}
									if fidderAB_bytes, err := msgpack.Marshal(fiddlerAB); err == nil {
										_, err := safe_instance.Invoke(ctx, k+"_tcp_fiddler_ab", fidderAB_bytes)
										if err := saveHandleMockErr("", []byte{}, []byte{}, startTime, index, "", "", "fiddler_ab ", err, uID); err != nil {
											fmt.Print("saveHandleMockErr ", err.Error())
										}
									}
								}
							}
						}
					}
					return true
				})
			}
		}
		response.Header.ReportId = uID
		util.JsonResponse(rw, response)
	}
}
func OauthLogin(rpcHandler RpcAble) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		if req.Body != nil {
			req.ParseForm()
			credential := req.Form.Get("credential")
			if email, secret, valid_token := Security.IsJWTValid(credential); valid_token {
				permission, er := rpcHandler.UserPermissionJWT(email, secret)
				if er == nil {
					if token, er := Security.ParseUserPermissionJWT(permission, secret); er == nil {
						Security.GiveToken(token, rw, req)
					}
				}
			}
		}

	}
}
func RegisterProcessors(cmdList []string, route string, rpcHandler RpcAble) {
	RpcHandler = rpcHandler
	registeredCommands = []string{}
	fmt.Println("RegisterProcessor cmdList", cmdList)
	for _, element := range cmdList {
		element_new := element
		if element_new == "" {
			continue
		}
		registeredCommands = append(registeredCommands, element_new)
		Processor := func(ctx context.Context, request, response interface{}) uint32 {
			requestNew := request
			source := rpcHandler.Source(ctx, requestNew)
			traceid := rpcHandler.TraceId(ctx, requestNew)
			startTime := time.Now()
			fmt.Println("hit-", element_new, "---", &hitcount, startTime)
			req_marshalling := &Bytes{}
			res_marshalling := &Bytes{}
			return CommonMockWasm(ctx, request, response, req_marshalling, res_marshalling, element_new, source, traceid)
		}
		rpcHandler.RegisterCommand(element_new+route, Processor)
	}
}
