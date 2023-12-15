package processor

import (
	"net/http"

	"github.com/wasmmock/wasm_mock_server/capabilities"
	"github.com/wasmmock/wasm_mock_server/model"
	"github.com/wasmmock/wasm_mock_server/util"
	proto "github.com/golang/protobuf/proto"
)

type mysqlWriterHTTP struct {
	loop_         string
	rw_           http.ResponseWriter
	response_     model.CallApiResponse
	mock_targets_ string
	code          []byte
}

func (c *mysqlWriterHTTP) loop() string {
	return c.loop_
}
func (c *mysqlWriterHTTP) rw() http.ResponseWriter {
	return c.rw_
}
func (c *mysqlWriterHTTP) response() model.CallApiResponse {
	return c.response_
}
func (c *mysqlWriterHTTP) mock_targets() string {
	return c.mock_targets_
}
func (c *mysqlWriterHTTP) jsonResponser(uID string, err error, errorCode uint32) {
	c.response_.Header.ReportId = uID
	if err != nil {
		c.response_.Header.Message = err.Error()
	}
	util.JsonResponse(c.rw_, c.response_)
}
func (c *mysqlWriterHTTP) wasmCode() ([]byte, error) {
	return c.code, nil
}

func (c *mysqlWriterHTTP) requester(command string, req interface{}, res interface{}) interface{} {
	switch v := req.(type) {
	case proto.Message:
		payload, _ := proto.Marshal(v)
		responseOb, _, _, httpErr := capabilities.HttpRequest(string(command), payload)
		if httpErr == nil {
			switch z := res.(type) {
			case proto.Message:
				proto.Unmarshal(responseOb, z)
			}
			return 0
		}
	}
	return 1
}
func (c *mysqlWriterHTTP) saveHandleInitErr(message string, err error, uID string) error {
	return saveHandleInitErr(c.rw_, &c.response_, message, err, uID)
}
func (c *mysqlWriterHTTP) saveHandleLoopErr(message string, err error, uID string) error {
	return saveHandleLoopErr(c.rw_, &c.response_, message, err, uID)
}

type mysqlWriterRPC struct {
	loop_         string
	rw_           http.ResponseWriter
	response_     model.CallApiResponse
	mock_targets_ string
	rpc_able      RpcAble
	code          []byte
}

func (c *mysqlWriterRPC) loop() string {
	return c.loop_
}
func (c *mysqlWriterRPC) rw() http.ResponseWriter {
	return c.rw_
}
func (c *mysqlWriterRPC) response() model.CallApiResponse {
	return c.response_
}
func (c *mysqlWriterRPC) mock_targets() string {
	return c.mock_targets_
}
func (c *mysqlWriterRPC) jsonResponser(uID string, err error, errorCode uint32) {
	c.response_.Header.ReportId = uID
	if err != nil {
		c.response_.Header.Message = err.Error()
	}
	util.JsonResponse(c.rw_, c.response_)
}
func (c *mysqlWriterRPC) wasmCode() ([]byte, error) {
	return c.code, nil
}
func (c *mysqlWriterRPC) requester(command string, req interface{}, res interface{}) interface{} {
	switch v := req.(type) {
	case proto.Message:
		payload, _ := proto.Marshal(v)
		responseOb, httpErr := capabilities.MysqlStatement(string(command), payload)
		if httpErr == nil {
			switch z := res.(type) {
			case proto.Message:
				proto.Unmarshal(responseOb, z)
			}
			return 0
		}

	}
	return 1
}
func (c *mysqlWriterRPC) saveHandleInitErr(message string, err error, uID string) error {
	return saveHandleInitErr(c.rw_, &c.response_, message, err, uID)
}
func (c *mysqlWriterRPC) saveHandleLoopErr(message string, err error, uID string) error {
	return saveHandleLoopErr(c.rw_, &c.response_, message, err, uID)
}
