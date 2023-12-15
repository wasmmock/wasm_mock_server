package processor

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wasmmock/wasm_mock_server/capabilities"
	"github.com/wasmmock/wasm_mock_server/tcpproxy"

	"github.com/wasmmock/wasm_mock_server/util"
	"github.com/vmihailenco/msgpack"
)

var hostCallSpecific = func(binding string, namespace string, operation string, payload []byte) ([]byte, error) {
	return []byte{}, nil
}

func RegisterHostCall(f func(string, string, string, []byte) ([]byte, error)) {
	hostCallSpecific = f
}
func hostCall(ctx context.Context, binding, namespace, operation string, payload []byte) ([]byte, error) {
	// Route the payload to any custom functionality accordingly.
	// You can even route to other waPC modules!!!
	switch namespace {
	case "foo":
		switch operation {
		case "echo":
			return payload, nil // echo
		case "get_index":
			//index := indexMap[strings.Split(binding, "?")[0]]
			index := indexMap[binding]
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(index))
			return b, nil
		case "assert_pass":
			//binding:endpoint
			e := util.Expectation{Pass: true, Description: string(payload)}
			if b := ctx.Value("uid"); b != nil {
				binding = b.(string)
			}
			fmt.Println("assert_pass uid", binding)
			if v := ctx.Value("report_index"); v != nil {
				if index, ok := v.(int); ok {
					safeReports.SetExpectation(binding, e, index)
				}
			} else {
				safeReports.AppendExpectation(binding, e)
			}
			return []byte{}, nil
		case "assert_fail":
			//binding:endpoint
			e := util.Expectation{Pass: false, Description: string(payload)}
			if b := ctx.Value("uid"); b != nil {
				binding = b.(string)
			}
			if v := ctx.Value("report_index"); v != nil {
				if index, ok := v.(int); ok {
					safeReports.SetExpectation(binding, e, index)
				}
			} else {
				safeReports.AppendExpectation(binding, e)
			}
			return []byte{}, nil
		case "step_pass":
			//binding:endpoint
			e := util.Step{Pass: true, Description: string(payload)}
			fmt.Println("step uid", binding)
			safeReports.AppendStep(binding, e)
			return []byte{}, nil
		case "step_fail":
			//binding:endpoint
			e := util.Step{Pass: false, Description: string(payload)}
			//uid := uidMap[binding]
			fmt.Println("step uid", binding)
			safeReports.AppendStep(binding, e)
			return []byte{}, nil
		case "error":
			//uid := uidMap[binding]
			e := util.Mock{Command: binding, Response: string(payload), Request: ""}
			safeReports.AppendMock(binding, e)
			return []byte{}, nil
		case "rpc_request":
			tpexReq2 := Bytes(payload)
			tpexRes := Bytes{}
			errCode := RpcHandler.RpcRequest(binding, &tpexReq2, &tpexRes)
			if errCode == 0 {
				return tpexRes, nil
			}
			return tpexRes, fmt.Errorf("%v", errCode)
		case "http_request":
			r, _, _, er := capabilities.HttpRequest(binding, payload)
			return r, er
		case "redis_get":
			return capabilities.RedisGetKey(binding, payload)
		case "redis_exist":
			return capabilities.RedisExistKey(binding, payload, true)
		case "redis_not_exist":
			return capabilities.RedisExistKey(binding, payload, false)
		case "redis_delete":
			return capabilities.RedisDeleteKey(binding, payload)
		case "memcache_get":
			return capabilities.MemcacheGetKey(binding, payload)
		case "memcache_exist":
			return capabilities.MemcacheExistKey(binding, payload, true)
		case "memcache_not_exist":
			return capabilities.MemcacheExistKey(binding, payload, false)
		case "memcache_delete":
			return capabilities.MemcacheDeleteKey(binding, payload)
		case "websocket":
			uid := strings.Split(binding, "|")[0]
			fn := strings.Split(binding, "|")[1]
			return WsMock(uid, fn, mockUINetConn, payload)
		case "websocket_call":
			uid := strings.Split(binding, "|")[0]
			fn := strings.Split(binding, "|")[1]
			index := strings.Split(binding, "|")[2]
			index_num, err := strconv.ParseInt(index, 10, 64)
			if err != nil {
				return []byte{}, err
			}
			return WsCall(uid, fn, index_num, callNetConn, payload)
		case "indexdb_get":
			fmt.Println("IndexDbGet", binding)
			return capabilities.IndexDbGet(&indexDBMap, binding)
		case "indexdb_store":
			fmt.Println("IndexDbStore", binding)
			return capabilities.IndexDbStore(&indexDBMap, binding, payload)
		case "mysql":
			fmt.Println("mysql request", binding)
			return capabilities.MysqlStatement(binding, payload)
		case "now":
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(time.Now().Unix()))
			return b, nil
		case "sleep":
			s := time.Duration(binary.LittleEndian.Uint64(payload))
			time.Sleep(s * time.Millisecond)
			return []byte{}, nil
		case "savefile":
			capabilities.SaveFile(payload, binding)
			return []byte{}, nil
		case "tcp_request":
			var item []capabilities.TcpReq
			err := msgpack.Unmarshal(payload, &item)
			if err != nil {
				e := util.Mock{Command: binding, Response: "", Request: err.Error()}
				safeReports.AppendMock(binding, e)
				return []byte{}, nil
			}
			tcpRes_s := tcpproxy.TcpRequest(item, binding)
			for _, tcpRes := range tcpRes_s {
				if tcpRes.ReportType == "Step" {
					s := util.Step{Pass: true, Description: tcpRes.String}
					safeReports.AppendStep(binding, s)
				} else {
					e := util.Mock{Command: binding, Response: tcpRes.String, Request: ""}
					safeReports.AppendMock(binding, e)
				}
			}
			return []byte{}, nil
		case "tcp_response":
			var items []capabilities.TcpReq
			err := msgpack.Unmarshal(payload, &items)
			if err != nil {
				fmt.Println("tcp_response err", err)
			}
			_ = err
			tcpproxy.TcpResponse(items, binding)
			if uID, ok := MockCommandUidMap.Get(binding); ok {
				for _, item := range items {
					s := util.Step{Pass: true, Description: item.String}
					safeReports.AppendStep(uID, s)
				}

			}
			return []byte{}, nil
		}
	case "selenium":
		switch operation {
		case "start":
			return capabilities.SeleniumStart(&wdMap, &wdServiceMap, binding)
		case "get":
			return capabilities.SeleniumGet(&wdMap, binding, payload)
		case "find_element":
			return capabilities.SeleniumFindElement(&wdMap, &wdElementMap, binding, payload)
		case "find_elements":
			bindingArr := strings.Split(binding, "|")
			if len(bindingArr) > 1 {
				uid := bindingArr[0]
				index, err := strconv.ParseInt(bindingArr[1], 10, 64)
				if err != nil {
					return []byte{}, err
				}
				return capabilities.SeleniumFindElements(&wdMap, &wdElementMap, uid, int(index), payload)
			}
			return []byte{}, fmt.Errorf("Cannot split the uid and the index for elements")
		case "click":
			return capabilities.SeleniumClick(&wdElementMap, binding)
		case "send_keys":
			return capabilities.SeleniumSendKeys(&wdElementMap, binding, payload)
		case "get_cookies":
			return capabilities.SeleniumGetCookies(&wdMap, binding)
		case "close":
			return capabilities.SeleniumClose(&wdMap, &wdServiceMap, &wdElementMap, binding)
		}

	}
	return hostCallSpecific(binding, namespace, operation, payload)
}
