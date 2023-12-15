package processor

import (
	"context"

	"github.com/wasmmock/wasm_mock_server/capabilities"
)

type TcpFiddleAB struct {
	ReqA        []byte
	ReqAString  string
	ResB        []byte
	ResBString  string
	ReportIndex int64
	End         bool
	PortMap     string
}

func TcpModifyRequest(req *Bytes, port_map string, l_remote_addr string, real_l_remote_addr string, real_r_local_addr string) (error, []capabilities.Entity) {
	//func(rw http.ResponseWriter, req *http.Request)
	if len(port_map) > 0 {
		response := &Bytes{}
		ctx := context.Background()
		//element_new := req.URL.Path
		request_index_in_report := int64(0)
		errcode, capEntity := TcpFiddlerMockWasmBeforeReq(ctx, req, response, port_map, &request_index_in_report, l_remote_addr, real_l_remote_addr, real_r_local_addr)
		_ = errcode
		return nil, capEntity

	}
	return nil, []capabilities.Entity{}
}

func TcpModifyResponse(res *Bytes, port_map string, l_remote_addr string, real_l_remote_addr string, real_r_local_addr string) (error, []capabilities.Entity) {
	//func(rw http.ResponseWriter, req *http.Request)
	if len(port_map) > 0 {
		//response := res
		ctx := context.Background()
		//element_new := req.URL.Path
		errcode, capEntity := TcpFiddlerMockWasmBeforeRes(ctx, res, port_map, false, l_remote_addr, real_l_remote_addr, real_r_local_addr)
		_ = errcode
		return nil, capEntity
	}

	return nil, []capabilities.Entity{}
}
