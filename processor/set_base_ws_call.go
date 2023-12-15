package processor

import (
	"io/ioutil"
	"net/http"

	"github.com/wasmmock/wasm_mock_server/model"
	"github.com/wasmmock/wasm_mock_server/util"
)

var baseWsCallCode []byte

func SetBaseWsCall(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)
	response.Header.Message = "ok"
	code, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	baseWsCallCode = code
	return
}
