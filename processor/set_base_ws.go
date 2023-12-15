package processor

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/wasmmock/wasm_mock_server/internal/box"
	"github.com/wasmmock/wasm_mock_server/model"
	"github.com/wasmmock/wasm_mock_server/util"
	wapc "github.com/wapc/wapc-go"
	"github.com/wapc/wapc-go/engines/wazero"
)

var baseWsMockModule wapc.Module

func SetBaseWsMock(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)
	response.Header.Message = "ok"
	code, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	ctx := context.Background()
	engine := wazero.Engine()
	module, err := engine.New(ctx, hostCall, code, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	//module, err := wapc.New(code, hostCall)
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	baseWsMockModule = module
	util.JsonResponse(rw, response)
	return
}
func SetBaseWsMockFromBox() {
	code := box.Get("/websocket_set_mock_new.wasm")
	ctx := context.Background()
	engine := wazero.Engine()
	module, err := engine.New(ctx, hostCall, code, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	//module, err := wapc.New(code, hostCall)
	if err != nil {
		fmt.Println("read websocket.wasm failed")
		return
	}
	baseWsMockModule = module
	return
}
func SetBaseWsCallFromBox() {
	baseWsCallCode = box.Get("/websocket_call_new.wasm")
	return
}
