package processor

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/wasmmock/wasm_mock_server/capabilities"
	"github.com/wasmmock/wasm_mock_server/util"
	proto "github.com/golang/protobuf/proto"
	"github.com/vmihailenco/msgpack"
)

func Max(x, y uint32) uint32 {
	if x < y {
		return y
	}
	return x
}

func CommonMockWasm(ctx context.Context, request, response, req_marshalling, res_marshalling interface{}, element_new string, source string, traceid string) uint32 {
	startTime := time.Now()
	fmt.Println("hit-", element_new, "---", &hitcount, startTime)
	newHits := []util.Hit{}
	hits.Range(func(j util.Hit) bool {
		if startTime.Sub(j.Time).Milliseconds() < 5000 {
			newHits = append(newHits, j)
		}
		return true
	})
	newHits = append(newHits, util.Hit{Time: startTime, Command: element_new, StartTime: startTime.Format("15:05:05")})
	for _, j := range newHits {
		hits.Store(j)
	}
	//hitcount = hitcount + 1
	atomic.AddInt32(&hitcount, 1)
	index := indexMap[element_new]
	returnCode := uint32(0)
	if returnCodeStr, er := indexDBMap.Get("ReturnCode|" + element_new); er == nil {
		if returnCodePar, er := strconv.ParseInt(string(returnCodeStr), 10, 64); er == nil {
			returnCode = uint32(returnCodePar)
		}
	}
	if mockUID, ok := MockCommandMockUidMap.Get(element_new); ok {
		if instance, ok := mockUidInstanceMap.Get(mockUID); ok {
			_, err3 := instance.Invoke(ctx, "save_command", []byte(element_new))
			if err3 != nil {
				fmt.Print("save_command ", err3)
			}
			r := request.(*Bytes)
			uID, ok := MockCommandUidMap.Get(element_new)
			payload, err := proto.Marshal(r)
			res := response.(*Bytes)

			if err != nil && !ok {
				fmt.Print("err ", err)
				return Max(1, returnCode)
			}
			result, err := instance.Invoke(ctx, element_new, payload)
			err = proto.Unmarshal(result, res)
			if !ok {
				return returnCode
			}
			//(command string, reqJSON []byte, startTime time.Time,index int64, source string, traceid string, what string, err error, uid string)
			if err := saveHandleMockErr(element_new, []byte{}, []byte{}, startTime, index, source, traceid, "mock request init marshal ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr ", err.Error())
				return Max(1, returnCode)
			}
			reqJSON, err := instance.Invoke(ctx, element_new+"_modify_req", payload)
			if err := saveHandleMockErr(element_new, reqJSON, []byte{}, startTime, index, source, traceid, "mock req invoke ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr ", err.Error())
				return Max(1, returnCode)
			} else {
				res := req_marshalling.(*Bytes)
				proto.Unmarshal(reqJSON, res)
			}
			_, err = instance.Invoke(ctx, "save_uid", []byte(uID))
			if err := saveHandleMockErr(element_new, reqJSON, []byte{}, startTime, index, source, traceid, "save uid ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr save uid", err.Error(), uID)
				return Max(1, returnCode)
			}
			if err := saveHandleMockErr(element_new, reqJSON, []byte{}, startTime, index, source, traceid, "mock command invoke ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr ", err.Error())
				return Max(1, returnCode)
			}
			endTime := time.Now()
			duration := util.DurationToString(endTime.Sub(startTime))
			err = proto.Unmarshal(result, res)
			if err := saveHandleMockErr(element_new, reqJSON, []byte{}, startTime, index, source, traceid, "mock res unmarshal ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr ", err)
				return Max(1, returnCode)
			}
			//modify_res marshalling
			resJSON, err := instance.Invoke(ctx, element_new+"_modify_res", result)
			if err := saveHandleMockErr(element_new, reqJSON, resJSON, startTime, index, source, traceid, "mock res _modify_res ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr ", err)
				return Max(1, returnCode)
			} else {
				res := res_marshalling.(*Bytes)
				proto.Unmarshal(resJSON, res)
			}
			mo := util.Mock{
				Command:  element_new,
				Response: string(resJSON),
				Request:  string(reqJSON),
				Duration: duration,
				Index:    index,
				Pass:     true,
				EndTime:  endTime.Format("15:05:05"),
				Source:   source,
				TraceId:  traceid,
			}
			//todo:all ReturnCode even if there is no reportuID
			safeReports.SetMock(uID, mo, int(index))

		} else {
			fmt.Println("Error mockUidInstanceMap", mockUidInstanceMap, mockUID, "command ", element_new)
		}
	} else {
		fmt.Println("Error MockCommandMockUidMap", MockCommandMockUidMap, "command", element_new)
	}

	return returnCode
}
func CommonMockWasmV2(ctx context.Context, type_ string, request, response, req_marshalling, res_marshalling interface{}, element_new string, source string, traceid string) uint32 {
	startTime := time.Now()
	fmt.Println("hit-", element_new, "---", &hitcount, startTime)
	newHits := []util.Hit{}
	hits.Range(func(j util.Hit) bool {
		if startTime.Sub(j.Time).Milliseconds() < 5000 {
			newHits = append(newHits, j)
		}
		return true
	})
	newHits = append(newHits, util.Hit{Time: startTime, Command: element_new, StartTime: startTime.Format("15:05:05")})
	for _, j := range newHits {
		hits.Store(j)
	}
	//hitcount = hitcount + 1
	atomic.AddInt32(&hitcount, 1)
	index := indexMap[element_new]
	returnCode := uint32(0)
	if returnCodeStr, er := indexDBMap.Get("ReturnCode|" + element_new); er == nil {
		if returnCodePar, er := strconv.ParseInt(string(returnCodeStr), 10, 64); er == nil {
			returnCode = uint32(returnCodePar)
		}
	}
	if mockUID, ok := MockCommandMockUidMap.Get(element_new); ok {
		if instance, ok := mockUidInstanceMap.Get(mockUID); ok {
			_, err3 := instance.Invoke(ctx, "save_command", []byte(element_new))
			if err3 != nil {
				fmt.Print("save_command ", err3)
			}
			r := request.(*Bytes)
			uID, ok := MockCommandUidMap.Get(element_new)
			payload, err := proto.Marshal(r)
			res := response.(*Bytes)

			if err != nil && !ok {
				fmt.Print("err ", err)
				return Max(1, returnCode)
			}
			result, err := instance.Invoke(ctx, element_new, payload)
			err = proto.Unmarshal(result, res)
			if !ok {
				return returnCode
			}
			//(command string, reqJSON []byte, startTime time.Time,index int64, source string, traceid string, what string, err error, uid string)
			if err := saveHandleMockErr(element_new, []byte{}, []byte{}, startTime, index, source, traceid, "mock request init marshal ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr ", err.Error())
				return Max(1, returnCode)
			}
			reqJSON, err := instance.Invoke(ctx, element_new+"_"+type_+"_modify_req", payload)
			if err := saveHandleMockErr(element_new, reqJSON, []byte{}, startTime, index, source, traceid, "mock req invoke ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr ", err.Error())
				return Max(1, returnCode)
			} else {
				res := req_marshalling.(*Bytes)
				proto.Unmarshal(reqJSON, res)
			}
			_, err = instance.Invoke(ctx, "save_uid", []byte(uID))
			if err := saveHandleMockErr(element_new, reqJSON, []byte{}, startTime, index, source, traceid, "save uid ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr save uid", err.Error(), uID)
				return Max(1, returnCode)
			}
			if err := saveHandleMockErr(element_new, reqJSON, []byte{}, startTime, index, source, traceid, "mock command invoke ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr ", err.Error())
				return Max(1, returnCode)
			}
			endTime := time.Now()
			duration := util.DurationToString(endTime.Sub(startTime))
			err = proto.Unmarshal(result, res)
			if err := saveHandleMockErr(element_new, reqJSON, []byte{}, startTime, index, source, traceid, "mock res unmarshal ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr ", err)
				return Max(1, returnCode)
			}
			//modify_res marshalling
			resJSON, err := instance.Invoke(ctx, element_new+"_"+type_+"_modify_res", result)
			if err := saveHandleMockErr(element_new, reqJSON, resJSON, startTime, index, source, traceid, "mock res _modify_res ", err, uID); err != nil {
				fmt.Print("saveHandleMockErr ", err)
				return Max(1, returnCode)
			} else {
				res := res_marshalling.(*Bytes)
				proto.Unmarshal(resJSON, res)
			}
			mo := util.Mock{
				Command:  element_new,
				Response: string(resJSON),
				Request:  string(reqJSON),
				Duration: duration,
				Index:    index,
				Pass:     true,
				EndTime:  endTime.Format("15:05:05"),
				Source:   source,
				TraceId:  traceid,
			}
			//todo:all ReturnCode even if there is no reportuID
			safeReports.SetMock(uID, mo, int(index))

		} else {
			fmt.Println("Error mockUidInstanceMap", mockUidInstanceMap, mockUID, "command ", element_new)
		}
	} else {
		fmt.Println("Error MockCommandMockUidMap", MockCommandMockUidMap, "command", element_new)
	}

	return returnCode
}
func FiddlerMockWasmBeforeReq(ctx context.Context, request, new_request interface{}, element_new string, url_path string, request_index_in_report *int64) (uint32, bool) {
	startTime := time.Now()
	fmt.Println("hit-", element_new, "---", &hitcount, startTime)
	newHits := []util.Hit{}
	hits.Range(func(j util.Hit) bool {
		if startTime.Sub(j.Time).Milliseconds() < 5000 {
			newHits = append(newHits, j)
		}
		return true
	})
	newHits = append(newHits, util.Hit{Time: startTime, Command: element_new, StartTime: startTime.Format("15:05:05")})
	for _, j := range newHits {
		hits.Store(j)
	}
	//hitcount = hitcount + 1
	atomic.AddInt32(&hitcount, 1)
	if mockUID, ok := MockCommandMockUidMap.Get(element_new); ok {
		if instance, ok := mockUidInstanceMap.Get(mockUID); ok {
			r := request.(*Bytes)
			payload, err := proto.Marshal(r)
			res := new_request.(*Bytes)
			if err != nil {
				fmt.Print("err ", err)
				return 1, false
			}
			fmt.Println("payload", string(payload))
			result, err := instance.Invoke(ctx, element_new+"_http_modify_req", payload)
			if err != nil {
				fmt.Println("_modify_req error", element_new, " ", err.Error())
			}
			err = proto.Unmarshal(result, res)
			uID, ok := MockCommandUidMap.Get(element_new)
			uniqueUrlPaths, ok2 := fiddlerUniqueUrlPath[uID]
			if ok && ok2 {
				var found = false
				for _, v := range uniqueUrlPaths {
					if v == url_path {
						found = true
					}
				}
				if !found {
					fiddlerUniqueUrlPath[uID] = append(fiddlerUniqueUrlPath[uID], url_path)
					reqJSON := result
					count := int64(safeReports.Len(uID))
					safeReports.AppendUnitTest(uID, count)
					iStep := util.Step{Pass: true, Description: url_path}
					safeReports.AppendStep(uID, iStep)
					safeReports.SetRequest(uID, string(reqJSON), count)

					*request_index_in_report = count
					return 0, false
				}

				return 0, true
			}
		} else {
			fmt.Println("Error mockUidInstanceMap", mockUidInstanceMap, mockUID, "command ", url_path)
		}
	} else {
		fmt.Println("Error MockCommandMockUidMap", MockCommandMockUidMap, "command", url_path)
	}
	return 0, false
}

type TcpPayload struct {
	Laddr     string
	Raddr     string
	Payload   string
	Tcp_Items []capabilities.TcpItem
}

func TcpFiddlerMockWasmBeforeReq(ctx context.Context, request, new_request interface{}, element_new string, request_index_in_report *int64, l_remote_addr string, real_l_remote_addr string, real_r_local_addr string) ([]uint32, []capabilities.Entity) {
	startTime := time.Now()
	fmt.Println("hit-", element_new, "---", &hitcount, startTime)
	if mockUID, ok := MockCommandMockUidMap.Get(element_new); ok {

		//if instance, ok := mockUidInstanceMap.Get(l_remote_addr); ok {
		if instance, ok := mockUidInstanceMap.Get(mockUID); ok {
			r := request.(*Bytes)
			payload, err := proto.Marshal(r)
			//res := new_request.(*Bytes)
			if err != nil {
				fmt.Print("err ", err)
				return []uint32{1}, []capabilities.Entity{}
			}
			fmt.Println("save_command... element_new", element_new)
			_, err3 := instance.Invoke(ctx, "save_command", []byte(element_new))
			if err3 != nil {
				fmt.Print("save_command ", err3)
			}
			fmt.Println("_modify_req... element_new", element_new, len(payload))
			tcp_payload := TcpPayload{
				Payload: b64.StdEncoding.EncodeToString(payload),
				Laddr:   real_l_remote_addr,
				Raddr:   real_r_local_addr,
			}
			p, _ := msgpack.Marshal(tcp_payload)
			result, err := instance.Invoke(ctx, element_new+"_modify_req", p)
			if err != nil {
				fmt.Println("after _modify_req... element_new err", element_new, err.Error())
			}
			fmt.Println("after _modify_req... element_new", element_new, len(p))
			if bytes.Equal(result, []byte("/continue")) {
				proto.Unmarshal(result, r)
				return []uint32{0}, []capabilities.Entity{capabilities.Entity{}}
			}
			if err != nil {
				return []uint32{1}, []capabilities.Entity{}
			}
			var items []capabilities.TcpItem
			err = msgpack.Unmarshal(result, &items)
			if err != nil {
				fmt.Println("msgpack err BeforeReq", err)
			}
			uID, _ := MockCommandUidMap.Get(element_new)
			var entity_arr []capabilities.Entity
			var errorcode_arr []uint32
			*r = []byte{}
			for _, item := range items {
				sDec, _ := b64.StdEncoding.DecodeString(item.Payload)
				*r = append(*r, sDec...)
				reqJSON := item.String
				count := int64(safeReports.Len(uID))
				if _, ok := tcpRequestIdMap[uID]; ok {
					if item.Id != "" {
						tcpRequestIdMap[uID][item.Id] = capabilities.TcpReq{
							Payload: b64.StdEncoding.EncodeToString(payload),
							String:  reqJSON,
							Index:   int(count),
						}
						safeReports.AppendUnitTest(uID, count)
						iStep := util.Step{Pass: true, Description: element_new}
						safeReports.AppendStep(uID, iStep)
						safeReports.SetRequest(uID, reqJSON, count)
						errorcode_arr = append(errorcode_arr, 0)
						entity_arr = append(entity_arr, capabilities.Entity{
							Id: item.Id,
						})
					}
				}
			}
			fmt.Println("total _modify_req... element_new", element_new, "payload len:", len(payload), "new request len:", len(*r))

			return errorcode_arr, entity_arr
		} else {
			fmt.Println("Error mockUidInstanceMap", mockUidInstanceMap, mockUID, "command ", element_new)
		}
	} else {
		fmt.Println("Error MockCommandMockUidMap", MockCommandMockUidMap, "command", element_new)
	}
	return []uint32{0}, []capabilities.Entity{}
}
func FiddlerMockWasmBeforeRes(ctx context.Context, request, response interface{}, element_new string, reqA *http.Request, request_index_in_report int64, skipReport bool) uint32 {
	if mockUID, ok := MockCommandMockUidMap.Get(element_new); ok {
		if instance, ok := mockUidInstanceMap.Get(mockUID); ok {
			r := request.(*Bytes)
			payload, err := proto.Marshal(r)
			res := response.(*Bytes)
			if err != nil {
				fmt.Print("err ", err)
				return 1
			}
			result, err := instance.Invoke(ctx, element_new+"_http_modify_res", payload)
			if err != nil {
				fmt.Println("after _modify_res... element_new err", element_new, err.Error())
			}
			err = proto.Unmarshal(result, res)

			uID, ok := MockCommandUidMap.Get(element_new)
			if ok && !skipReport {
				resJSON := result
				safeReports.SetResponse(uID, string(resJSON), int(request_index_in_report))
				resB := response.(*Bytes)
				res_b, _ := proto.Marshal(resB)
				wasmRes := &capabilities.Response{}
				json.Unmarshal(res_b, wasmRes)
				go func() {
					channel := fiddlerQueue[uID]
					channel <- FiddleAB{
						ReqA:        reqA,
						ResB:        *wasmRes,
						ReportIndex: request_index_in_report,
					}
				}()
				return 0
			}
		} else {
			fmt.Println("Error mockUidInstanceMap", mockUidInstanceMap, mockUID, "command ", reqA.URL.Path)
		}
	} else {
		fmt.Println("Error MockCommandMockUidMap", MockCommandMockUidMap, "command", reqA.URL.Path)
	}
	return 0
}

func TcpFiddlerMockWasmBeforeRes(ctx context.Context, response interface{}, element_new string, skipReport bool, l_remote_addr string, real_l_remote_addr string, real_r_local_addr string) ([]uint32, []capabilities.Entity) {
	if mockUID, ok := MockCommandMockUidMap.Get(element_new); ok {
		//if instance, ok := mockUidInstanceMap.Get(l_remote_addr); ok {
		if instance, ok := mockUidInstanceMap.Get(mockUID); ok {
			r := response.(*Bytes)
			payload, err := proto.Marshal(r)
			if err != nil {
				fmt.Print("err ", err)
				return []uint32{1}, []capabilities.Entity{capabilities.Entity{}}
			}
			_, err3 := instance.Invoke(ctx, "save_command", []byte(element_new))
			if err3 != nil {
				fmt.Print("save_command ", err3)
			}
			tcp_payload := TcpPayload{
				Payload: b64.StdEncoding.EncodeToString(payload),
				Laddr:   real_l_remote_addr,
				Raddr:   real_r_local_addr,
			}
			p, _ := msgpack.Marshal(tcp_payload)
			result, err := instance.Invoke(ctx, element_new+"_modify_res", p)
			if err != nil {
				//return 0
				fmt.Println("result, error", err)
				return []uint32{1}, []capabilities.Entity{capabilities.Entity{}}
			}
			if bytes.Equal(result, []byte("/continue")) {
				proto.Unmarshal(result, r)
				return []uint32{0}, []capabilities.Entity{capabilities.Entity{}}
			}
			var items []capabilities.TcpItem
			err = msgpack.Unmarshal(result, &items)
			if err != nil {
				fmt.Println(" TcpFiddlerMockWasmBeforeRes msgpack err", err)
				return []uint32{1}, []capabilities.Entity{capabilities.Entity{}}
			}
			*r = []byte{}
			var entity_arr []capabilities.Entity
			var errorcode_arr []uint32
			if len(items) > 1 {
				log.Println("consolidated more than 1", items)
			}
			for _, item := range items {
				sDec, _ := b64.StdEncoding.DecodeString(item.Payload)
				*r = append(*r, sDec...)
				uID, _ := MockCommandUidMap.Get(element_new)
				resJSON := item.String
				if _, ok := tcpRequestIdMap[uID]; ok {
					if item.Id != "" {
						if v, ok := tcpRequestIdMap[uID][item.Id]; ok {
							safeReports.SetResponse(uID, resJSON, int(v.Index))
							fmt.Println(" TcpFiddlerMockWasmBeforeRes queue")
							go func() {
								sDec2, _ := b64.StdEncoding.DecodeString(v.Payload)
								channel := tcpFiddlerQueue[uID]
								channel <- TcpFiddleAB{
									ReqA:        sDec2,
									ReqAString:  v.String,
									ResB:        sDec,
									ResBString:  resJSON,
									ReportIndex: int64(v.Index),
									PortMap:     element_new,
								}
							}()
						}
					}
				}
				errorcode_arr = append(errorcode_arr, uint32(0))
				entity_arr = append(entity_arr, capabilities.Entity{})
			}
			return errorcode_arr, entity_arr
		} else {
			fmt.Println("Error mockUidInstanceMap", mockUidInstanceMap, mockUID, "command ", element_new)
		}
	} else {
		fmt.Println("Error MockCommandMockUidMap", MockCommandMockUidMap, "command", element_new)
	}
	return []uint32{0}, []capabilities.Entity{capabilities.Entity{}}
}
