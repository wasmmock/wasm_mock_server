package main

import (
	"context"

	"github.com/golang/protobuf/proto"
	processor "github.com/wasmmock/wasm_mock_server/processor"
	"github.com/wasmmock/wasm_mock_server/server"
)

type LocalRpcAble struct {
}

func (v LocalRpcAble) Start(s []string) {
}
func (v LocalRpcAble) Stop() {
}
func (v LocalRpcAble) HostCallSpecific(a string, b string, c string, buffer []byte) ([]byte, error) {
	return []byte{}, nil
}
func (v LocalRpcAble) RegisterCommand(cmd string, callback func(ctx context.Context, request, response interface{}) uint32) {
	return
}
func (v LocalRpcAble) RpcRequest(cmd string, req proto.Message, res proto.Message) uint32 {
	return 0
}
func (v LocalRpcAble) Source(ctx_a context.Context, key interface{}) string {
	return ""
}
func (v LocalRpcAble) TraceId(ctx_a context.Context, key interface{}) string {
	return ""
}
func (v LocalRpcAble) FiddleQueue(chan_a chan processor.FiddleAB, chan_b chan processor.FiddleAB) {
	for v := range chan_a {
		chan_b <- v
	}
}
func (v LocalRpcAble) TcpFiddleQueue(chan_a chan processor.TcpFiddleAB, chan_b chan processor.TcpFiddleAB) {
	for v := range chan_a {
		chan_b <- v
	}
}
func (v LocalRpcAble) UserPermissionJWT(a string, b string) (string, error) {
	return "", nil
}
func main() {
	lc := LocalRpcAble{}
	server.NewServer(lc, "")
}
