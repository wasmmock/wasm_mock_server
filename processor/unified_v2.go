package processor

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/audiolion/ipip"
	proto "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	wapc "github.com/wapc/wapc-go"
	"github.com/wapc/wapc-go/engines/wazero"
	"github.com/wasmmock/wasm_mock_server/capabilities"
	"github.com/wasmmock/wasm_mock_server/model"
	Security "github.com/wasmmock/wasm_mock_server/security"
	"github.com/wasmmock/wasm_mock_server/tcpproxy"
	"github.com/wasmmock/wasm_mock_server/util"
)

func HandleAT(loop string, code []byte, isHttp bool, isHttpFiddlerAB bool, http_targetList []string, response *model.CallApiResponse, rw http.ResponseWriter, req *http.Request, rpcHandler RpcAble) {
	if isHttpFiddlerAB == false {
		if loop != "0" {
			req.Body = ioutil.NopCloser(bytes.NewBuffer(code))
			CallV2(rw, req)
		} else {
			util.JsonResponse(rw, response)
		}
	} else {
		HandleHttpFiddlerAB(code, isHttpFiddlerAB, http_targetList, response, rw, req, rpcHandler)
	}
}
func HandleHttpFiddlerAB(code []byte, isHttpFiddlerAB bool, http_targetList []string, response *model.CallApiResponse, rw http.ResponseWriter, req *http.Request, rpcHandler RpcAble) {
	if isHttpFiddlerAB {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(code))
		mock_targets := strings.Join(http_targetList, ",")
		if len(mock_targets) > 0 {
			req.URL.RawQuery += "&mock_targets=" + mock_targets
		}
		CallFiddler(rpcHandler)(rw, req)
	} else {
		util.JsonResponse(rw, response)
	}
}

func UnifiedV2(rpcHandler RpcAble) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		response := new(model.CallApiResponse)
		response.Header.Message = "ok"
		code, err := ioutil.ReadAll(req.Body)
		fmt.Println("raw query", req.URL.RawQuery)
		defer req.Body.Close()
		if err != nil {
			response.Header.Message = err.Error()
			util.JsonResponse(rw, response)
			return
		}
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

		var instance wapc.Instance
		if wasmLang == "go" {
			instance, err = module.Instantiate(ctx)
		} else {
			instance, err = module.Instantiate(ctx)
		}
		if err != nil {
			response.Header.Message = err.Error()
			util.JsonResponse(rw, response)
			return
		}

		added_fn, err := instance.Invoke(ctx, "add_functions", []byte{})

		targets := string(added_fn)
		fmt.Println("added_fn", targets)
		if err == nil {
			response.Header.Message = "added_fn " + targets
		} else {
			response.Header.Message = err.Error()
			fmt.Println("added_fn err", err.Error())
		}
		loop_bytes, _ := instance.Invoke(ctx, "loop", []byte{})
		loop := "0"
		z := string(loop_bytes)
		l := strings.Split(z, "_")
		if len(l) > 1 {
			loop = l[1]
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
		targetList := strings.Split(targets, ",")
		var tcp_targetList = []string{}
		var http_targetList = []string{}
		var isHttp = false
		var isHttpFiddlerAB = false
		for _, target := range targetList {
			if strings.Contains(target, "_http_modify_req") || strings.Contains(target, "_http_modify_res") {
				isHttp = true
				t := strings.ReplaceAll(target, "_http_modify_req", "")
				t = strings.ReplaceAll(t, "_http_modify_res", "")
				var contain = false
				for _, http_target := range http_targetList {
					if http_target == t {
						contain = true
					}
				}
				if contain {
					//break
					continue
				}
				MockCommandMockUidMap.Store(t, uID)
				http_targetList = append(http_targetList, t)
				fiddlerBeforeRequestMap.Store(t, uID)
			} else if strings.Contains(target, "_tcp_modify_req") || strings.Contains(target, "_tcp_modify_res") {
				port_map := strings.ReplaceAll(target, "_tcp_modify_req", "")
				port_map = strings.ReplaceAll(port_map, "_tcp_modify_res", "")
				var contain = false
				for _, tcp_target := range tcp_targetList {
					if tcp_target == port_map {
						contain = true
					}
				}
				if contain {
					//break
					continue
				}
				MockCommandMockUidMap.Store(target, uID)
				fiddlerBeforeRequestMap.Store(target, uID)
			} else if strings.Contains(target, "_http_fiddler_ab") {
				isHttpFiddlerAB = true
			}
		}
		for _, port_map := range tcp_targetList {
			port_map_arr := strings.Split(port_map, "-:")
			if len(port_map_arr) < 2 {
				response.Header.Message = port_map + ": portmap should be in 3335-:3334 format"
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
		instanceref := util.NewSafeInstance(instance)
		mockUidInstanceMap.Set(uID, &instanceref)
		HandleAT(loop, code, isHttp, isHttpFiddlerAB, http_targetList, response, rw, req, rpcHandler)
		return
	}
}
