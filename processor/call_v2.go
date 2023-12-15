package processor

import (
	"io/ioutil"
	"net/http"

	"github.com/wasmmock/wasm_mock_server/model"
	"github.com/wasmmock/wasm_mock_server/util"
)

func CallV2(rw http.ResponseWriter, req *http.Request) {
	res := model.CallApiResponse{
		Header: model.Header{},
	}
	code, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		util.JsonResponse(rw, res)
	}

	wasmLangArr := req.URL.Query()["wasm_lang"]
	var wasmLang string = "rust"
	if len(wasmLangArr) > 0 {
		wasmLang = wasmLangArr[0]
	}
	loop := ""
	if len(req.URL.Query()["loop"]) > 0 {
		loop = req.URL.Query()["loop"][0]
	}
	wsWriter_ := httpWriterHTTP{
		loop_:         loop,
		rw_:           rw,
		mock_targets_: "",
		response_:     res,
		code:          code,
		wasm_lang_:    wasmLang,
	}
	CallWasmV2(&wsWriter_, "")
	return
}
