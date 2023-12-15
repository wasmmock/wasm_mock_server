package processor

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/wasmmock/wasm_mock_server/capabilities"
	"github.com/wasmmock/wasm_mock_server/model"
	"github.com/wasmmock/wasm_mock_server/util"
	"github.com/wasmmock/wasm_mock_server/util/postman"
	proto "github.com/golang/protobuf/proto"
)

type httpWriterHTTP struct {
	loop_               string
	rw_                 http.ResponseWriter
	response_           model.CallApiResponse
	mock_targets_       string
	wasm_lang_          string
	code                []byte
	previousCurlCommand string
	CurlCommand         util.CurlCommand
	PostmanItems        []postman.Item
}

func (c *httpWriterHTTP) loop() string {
	return c.loop_
}
func (c *httpWriterHTTP) rw() http.ResponseWriter {
	return c.rw_
}
func (c *httpWriterHTTP) response() model.CallApiResponse {
	return c.response_
}
func (c *httpWriterHTTP) mock_targets() string {
	return c.mock_targets_
}
func (c *httpWriterHTTP) wasmLang() string {
	return c.wasm_lang_
}
func (c *httpWriterHTTP) jsonResponser(uID string, err error, errorCode uint32) {
	c.response_.Header.ReportId = uID
	if err != nil {
		c.response_.Header.Message = err.Error()
	}
	util.JsonResponse(c.rw_, c.response_)
}
func (c *httpWriterHTTP) wasmCode() ([]byte, error) {
	return c.code, nil
}

func (c *httpWriterHTTP) requester(command string, req interface{}, res interface{}) interface{} {
	switch v := req.(type) {
	case proto.Message:
		payload, _ := proto.Marshal(v)
		c.CurlCommand = util.CurlCommand{}
		responseOb, curlCommand, postmanItem, httpErr := capabilities.HttpRequest(command, payload)
		c.PostmanItems = append(c.PostmanItems, postmanItem)

		c.CurlCommand = curlCommand
		wasmReq := capabilities.Request{}
		err := json.Unmarshal(payload, &wasmReq)
		if err == nil {
			wasmReqForReport := capabilities.RequestForReport{
				Http1x:   wasmReq.Http1x,
				HttpBody: string(wasmReq.HttpBody),
				ProxyUrl: wasmReq.ProxyUrl,
			}
			b, err1 := json.Marshal(wasmReqForReport)
			if err1 == nil {
				proto.Unmarshal(b, v)
			}
		}
		if httpErr == nil {
			switch z := res.(type) {
			case proto.Message:
				proto.Unmarshal(responseOb, z)
			}
			return 0
		} else {
			return httpErr
		}
	}
	return 1
}
func (c *httpWriterHTTP) curlCommand() string {
	return c.CurlCommand.String()
}
func (c *httpWriterHTTP) postman() (postman.Item, error) {
	if len(c.PostmanItems) > 0 {
		return c.PostmanItems[0], nil
	}
	return postman.Item{}, errors.New("no postman")
}
func (c *httpWriterHTTP) saveHandleInitErr(message string, err error, uID string) error {
	return saveHandleInitErr(c.rw_, &c.response_, message, err, uID)
}
func (c *httpWriterHTTP) saveHandleLoopErr(message string, err error, uID string) error {
	return saveHandleLoopErr(c.rw_, &c.response_, message, err, uID)
}

type httpWriterRPC struct {
	loop_         string
	rw_           http.ResponseWriter
	response_     model.CallApiResponse
	mock_targets_ string
	wasm_lang_    string
	rpc_able      RpcAble
	code          []byte
}

func (c *httpWriterRPC) loop() string {
	return c.loop_
}
func (c *httpWriterRPC) rw() http.ResponseWriter {
	return c.rw_
}
func (c *httpWriterRPC) response() model.CallApiResponse {
	return c.response_
}
func (c *httpWriterRPC) mock_targets() string {
	return c.mock_targets_
}
func (c *httpWriterRPC) wasmLang() string {
	return c.wasm_lang_
}
func (c *httpWriterRPC) jsonResponser(uID string, err error, errorCode uint32) {
	c.response_.Header.ReportId = uID
	if err != nil {
		c.response_.Header.Message = err.Error()
	}
	util.JsonResponse(c.rw_, c.response_)
}
func (c *httpWriterRPC) wasmCode() ([]byte, error) {
	return c.code, nil
}
func (c *httpWriterRPC) requester(command string, req interface{}, res interface{}) interface{} {
	switch v := req.(type) {
	case proto.Message:
		switch z := res.(type) {
		case proto.Message:
			return c.rpc_able.RpcRequest(command, v, z)
		}
	}
	return 1
}
func (c *httpWriterRPC) curlCommand() string {
	return ""
}
func (c *httpWriterRPC) postman() (postman.Item, error) {
	return postman.Item{}, nil
}
func (c *httpWriterRPC) saveHandleInitErr(message string, err error, uID string) error {
	return saveHandleInitErr(c.rw_, &c.response_, message, err, uID)
}
func (c *httpWriterRPC) saveHandleLoopErr(message string, err error, uID string) error {
	return saveHandleLoopErr(c.rw_, &c.response_, message, err, uID)
}
