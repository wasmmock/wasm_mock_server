package processor

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wasmmock/wasm_mock_server/model"
	Security "github.com/wasmmock/wasm_mock_server/security"
	"github.com/wasmmock/wasm_mock_server/tcpproxy"
	"github.com/wasmmock/wasm_mock_server/util/postman"

	"github.com/wasmmock/wasm_mock_server/capabilities"
	"github.com/wasmmock/wasm_mock_server/util"
	"github.com/audiolion/ipip"
	proto "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}
var mockUINetConn util.SafeWsConnMap = util.NewSafeWsConnMap()
var callNetConn util.SafeWsConnMap = util.NewSafeWsConnMap()
var traceNetConn util.SafeWsConnMap = util.NewSafeWsConnMap()

type wsWriterRPC struct {
	loop_         string
	rw_           http.ResponseWriter
	response_     model.CallApiResponse
	mock_targets_ string
	rpc_able      RpcAble
	code          []byte
	wasm_lang_    string
	ws            websocket.Conn
}

func (c *wsWriterRPC) loop() string {
	return c.loop_
}
func (c *wsWriterRPC) rw() http.ResponseWriter {
	return c.rw_
}
func (c *wsWriterRPC) response() model.CallApiResponse {
	return c.response_
}
func (c *wsWriterRPC) mock_targets() string {
	return c.mock_targets_
}
func (c *wsWriterRPC) wasmLang() string {
	return c.wasm_lang_
}
func (c *wsWriterRPC) jsonResponser(uID string, err error, errorCode uint32) {
}
func (c *wsWriterRPC) wasmCode() ([]byte, error) {
	return c.code, nil
}
func (c *wsWriterRPC) requester(command string, req interface{}, res interface{}) interface{} {
	switch v := req.(type) {
	case proto.Message:
		switch z := res.(type) {
		case proto.Message:
			return c.rpc_able.RpcRequest(command, v, z)
		}
	}
	return 1
}
func (c *wsWriterRPC) curlCommand() string {
	return ""
}
func (c *wsWriterRPC) postman() (postman.Item, error) {
	return postman.Item{}, nil
}
func (c *wsWriterRPC) saveHandleInitErr(message string, err error, uid string) error {
	if err != nil {
		if safeReports.Len(uid) == 0 {
			safeReports.AppendUnitTest(uid, 0)
		}
		c.response_.Header = model.Header{
			Message:  message + " " + err.Error(),
			ReportId: uid,
		}
		util.JsonResponseWS(&c.ws, c.response_)
		iStep := util.Step{Pass: false, Description: message + " " + err.Error()}
		safeReports.AppendStep(uid, iStep)
		safeReports.AppendEnd(uid, backgroundMock, hits.Clone())
	}
	return err
}
func (c *wsWriterRPC) saveHandleLoopErr(message string, err error, uid string) error {
	if err != nil {
		iStep := util.Step{Pass: false, Description: message + " " + err.Error()}
		safeReports.AppendStep(uid, iStep)
		safeReports.AppendEnd(uid, backgroundMock, hits.Clone())
	}
	return err
}

type wsResponseWriter struct {
}

func (c *wsResponseWriter) Header() http.Header {
	return make(map[string][]string)
}
func (c *wsResponseWriter) Write(paylaod []byte) (int, error) {
	return 200, nil
}
func (c *wsResponseWriter) WriteHeader(statusCode int) {

}
func WebsocketRPC(rpcHandler RpcAble) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		uID := uuid.New().String()
		// upgrader.CheckOrigin = func(r *http.Request) bool {
		// 	return Security.X_Api_Key(r)
		// }
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		c, err := upgrader.Upgrade(rw, req, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		ws_con := util.NewSafeWsConn(c)
		callNetConn.Set(uID, &ws_con)
		res := model.CallApiResponse{}
		//for {
		var k = capabilities.WsCallProtocol{}
		err = c.ReadJSON(&k)
		if err == nil {
			if k.Loop != "" {
				wsWriter_ := wsWriterRPC{
					loop_:         k.Loop,
					rw_:           &wsResponseWriter{},
					response_:     res,
					mock_targets_: k.Targets,
					code:          baseWsCallCode,
					rpc_able:      rpcHandler,
					wasm_lang_:    "rust",
				}
				CallWasm(&wsWriter_, uID) //uID= ws_uid
			}
			// } else {
			// 	break
			// }
		}
		return
	}
}
func WebsocketSetMock(rw http.ResponseWriter, req *http.Request) {
	uID := uuid.New().String()
	// upgrader.CheckOrigin = func(r *http.Request) bool {
	// 	return Security.X_Api_Key(r)
	// }
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	c, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	_, targets, err := c.ReadMessage()
	targetList := strings.Split(string(targets), ",")
	//instance, err := baseWsMockModule2.Instantiate("wasi_unstable")
	ctx := context.Background()
	instance, err := baseWsMockModule.Instantiate(ctx)
	instance.Invoke(ctx, "add_functions", []byte{})
	instance.Invoke(ctx, "save_ws_uid", []byte(uID))
	for _, target := range targetList {
		MockCommandMockUidMap.Store(target, uID)
		_, err := instance.Invoke(ctx, "add_ws_functions", []byte(target))
		if err != nil {
			fmt.Println("add_ws_functions fails ", target, err.Error(), time.Now())
		}
	}
	c.SetCloseHandler(func(code int, text string) error {
		mockUINetConn.Delete(uID)
		return nil
	})
	ws_con := util.NewSafeWsConn(c)
	mockUINetConn.Set(uID, &ws_con)

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
}
func WebsocketSetMockTcpFiddler(rw http.ResponseWriter, req *http.Request) {
	// upgrader.CheckOrigin = func(r *http.Request) bool {
	// 	return Security.X_Api_Key(r)
	// }
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	c, err := upgrader.Upgrade(rw, req, nil)

	var uID string
	uIDContext := req.Context().Value(Security.UIDKey)

	if uIDString, ok := uIDContext.(string); ok {
		uID = uIDString
	}
	if uIDContext == nil {
		uID = uuid.New().String()
		fmt.Println("WebsocketSetMockTcpFiddler uID error")
	}
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	_, targets, err := c.ReadMessage()

	targetList := strings.Split(string(targets), ",")
	tcpproxy.TcpMockTeardown(targetList, &MockCommandMockUidMap)
	c.SetCloseHandler(func(code int, text string) error {
		fmt.Println("c.SetCloseHandler")
		tcpproxy.TcpMockTeardown(targetList, &MockCommandMockUidMap)
		mockUINetConn.Delete(uID)
		Security.X_Api_Key_Map.Delete(uID)
		return nil
	})
	ws_con := util.NewSafeWsConn(c)
	mockUINetConn.Set(uID, &ws_con)
	for _, port_map := range targetList {
		port_map_arr := strings.Split(port_map, "-:")
		if len(port_map_arr) < 2 {
			return
		}
		localAddr := port_map_arr[0]
		if _, err := strconv.Atoi(localAddr); err == nil {
			localAddr = ":" + localAddr
		}
		laddr, err := net.ResolveTCPAddr("tcp", localAddr)
		if err != nil {
			log.Println("laddr connect error", localAddr, err)
			return
		}
		remoteAddr := port_map_arr[1]
		if _, err := strconv.Atoi(remoteAddr); err == nil {
			remoteAddr = ":" + remoteAddr
		}
		//unwrapTLS := false
		raddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
		if err != nil {
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
						wp := capabilities.WsCallProtocol{
							Fn:      "error",
							Payload: string(v.ErrorMsg),
						}
						ws_con.WriteJSON(wp)
						ws_con.Close()
						return
					}
				}
			}
		}
		Before_req := func(payload []byte, port_map, l_remote_addr string, real_lconn_addr string, real_r_local_addr string) ([]byte, []capabilities.Entity) {
			request := Bytes(payload)
			_, e := TcpModifyRequest(&request, port_map, l_remote_addr, real_lconn_addr, real_r_local_addr)
			res, _ := proto.Marshal(&request)
			return res, e
		}
		Before_res := func(payload []byte, port_map, l_remote_addr string, real_lconn_addr string, real_r_local_addr string) ([]byte, []capabilities.Entity) {
			request := Bytes(payload)
			_, e := TcpModifyResponse(&request, port_map, l_remote_addr, real_lconn_addr, real_r_local_addr)
			res, _ := proto.Marshal(&request)
			return res, e
		}
		port_map_c := port_map
		tcpproxy.PoolInit(uID, port_map_c, laddr, raddr, Before_req, Before_res, &MockCommandMockUidMap, &mockUidInstanceMap, baseWsMockModule)
	}

}
func WebsocketSetMockBeforeRequestFiddler(rw http.ResponseWriter, req *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	c, err := upgrader.Upgrade(rw, req, nil)
	uID := uuid.New().String()
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	_, targets, err := c.ReadMessage()

	targetList := strings.Split(string(targets), ",")
	for _, target := range targetList {
		MockCommandMockUidMap.Store(target, uID)
		fiddlerBeforeRequestMap.Store(target, uID)
	}
	ctx := context.Background()
	instance, err := baseWsMockModule.Instantiate(ctx)
	instance.Invoke(ctx, "add_functions", []byte{})
	instance.Invoke(ctx, "save_ws_uid", []byte(uID))
	for _, target := range targetList {
		MockCommandMockUidMap.Store(target, uID)
		_, err := instance.Invoke(ctx, "add_ws_functions", []byte(target))
		if err != nil {
			fmt.Println("add_ws_functions fails ", target, err.Error(), time.Now())
		}
	}
	c.SetCloseHandler(func(code int, text string) error {
		mockUINetConn.Delete(uID)
		return nil
	})
	ws_con := util.NewSafeWsConn(c)
	mockUINetConn.Set(uID, &ws_con)

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
}
func WebsocketTrace(rw http.ResponseWriter, req *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	uID := uuid.New().String()
	ws_con := util.NewSafeWsConn(c)
	targets := "/"
	if len(req.URL.Query()["targets"]) > 0 {
		targets = req.URL.Query()["targets"][0]
		traceUidMap.Store(uID, targets)
	}
	traceNetConn.Set(uID, &ws_con)
	c.SetCloseHandler(func(code int, text string) error {
		traceNetConn.Delete(uID)
		traceUidMap.Delete(uID)
		return nil
	})
}
func WsCall(uid string, fn string, index int64, callNetConn util.SafeWsConnMap, req []byte) ([]byte, error) {
	message := []byte{}
	err := fmt.Errorf("cannot find uid in callNetConn %v", uid)
	if c, ok := callNetConn.Get(uid); ok {
		wp := capabilities.WsCallProtocol{
			Fn:      fn,
			Payload: string(req),
			Index:   index,
		}
		fmt.Println("wscall wp", wp)
		err = c.WriteJSON(wp)

		if err != nil {
			fmt.Println("wscall err write", err)
			return []byte{}, err
		}

		ctx := context.Background()
		for {
			//c := callNetConn[uid]
			var k = capabilities.WsCallProtocol{}

			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println("WsCall err", err)
				break
			}
			if err2 := json.Unmarshal(message, &k); err2 == nil {
				if k.Binding != "" && k.Fn != "" && k.Payload != "" {
					p, _ := b64.StdEncoding.DecodeString(k.Payload)
					hostCall(ctx, k.Binding, "foo", k.Fn, p)
				} else {
					break
				}
			} else {
				//fmt.Println("err json unmarshal", err2, "messgage", string(message))
				break
			}
		}

	}
	return message, err
}
func WsMock(uid string, fn string, mockUINetConn util.SafeWsConnMap, req []byte) ([]byte, error) {
	if c, ok := mockUINetConn.Get(uid); ok {
		wp := capabilities.WsProtocol{
			Fn:      fn,
			Payload: b64.StdEncoding.EncodeToString(req),
			//Payload: string(req),
		}
		if msg, err := json.Marshal(wp); err == nil {
			if err := c.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				fmt.Println("WriteJSON ErrClosedPipe")
				var mockCommandMockUidMap_value_arr = make(map[string]string)
				MockCommandMockUidMap.Range(func(key string, value string) bool {
					if value == uid {
						mockCommandMockUidMap_value_arr[key] = value
						return false
					}
					return true
				})
				for key, _ := range mockCommandMockUidMap_value_arr {
					tcpproxy.TcpMockTeardown([]string{key}, &MockCommandMockUidMap)
					MockCommandMockUidMap.Delete(key)
					//Security.X_Api_Key_Map.Delete(key)
				}
				c.Close()
				return []byte{}, err
			}
		} else {
			return []byte{}, err
		}
		_, message, err2 := c.ReadMessage()
		if err2 != nil {
			fmt.Println("read ErrClosedPipe")
			var mockCommandMockUidMap_value_arr = make(map[string]string)
			MockCommandMockUidMap.Range(func(key string, value string) bool {
				if value == uid {
					mockCommandMockUidMap_value_arr[key] = value
					return false
				}
				return true
			})
			for key, _ := range mockCommandMockUidMap_value_arr {
				tcpproxy.TcpMockTeardown([]string{key}, &MockCommandMockUidMap)
				MockCommandMockUidMap.Delete(key)
				//Security.X_Api_Key_Map.Delete(key)
			}
			c.Close()
			return []byte{}, err2
		}

		return message, err2
	}
	return []byte{}, errors.New("can't find mockUINetConn uid")
}
func WsTrace(resp *http.Response, traceNetConn *util.SafeWsConnMap, readable_req_body string, readable_res_body string) {

	len := traceNetConn.SizeHint()

	if len > 0 {
		to_delete := []string{}
		trace := util.ConvertToTrace(resp, readable_req_body, readable_res_body)

		for uid, c := range traceNetConn.GetAll() {
			if targets, ok := traceUidMap.Get(uid); ok {
				targets_arr := strings.Split(targets, ",")
				var matches = false
				for _, t := range targets_arr {
					if strings.Contains(t, `\`) {
						var validID = regexp.MustCompile(t)
						if validID.MatchString(resp.Request.URL.Path) {
							matches = true
						}
					} else if strings.Contains(resp.Request.URL.Path, t) {
						matches = true
					}
				}
				if matches {
					if msg, err := json.Marshal(trace); err == nil {
						if err := c.WriteMessage(websocket.BinaryMessage, msg); err != nil {
							log.Println("WsTrace write error", err.Error())
							to_delete = append(to_delete, uid)
						}
					}
				}

			}

		}
		for _, v := range to_delete {
			traceNetConn.Delete(v)
		}
	}

}
