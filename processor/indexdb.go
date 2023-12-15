package processor

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/wasmmock/wasm_mock_server/model"
	"github.com/wasmmock/wasm_mock_server/util"
)

type Index_db_struct struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func StoreIndexDB(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)
	response.Header.Message = "ok"
	var t Index_db_struct
	body, err := ioutil.ReadAll(req.Body)
	json.Unmarshal(body, &t)
	defer req.Body.Close()
	if err != nil {
		response.Header.Message = err.Error()
		util.JsonResponse(rw, response)
		return
	}
	indexDBMap.Store(t.Key, []byte(t.Value))
	util.JsonResponse(rw, response)
	return
}
func GetIndexDB(rw http.ResponseWriter, req *http.Request) {
	response := new(model.CallApiResponse)

	response.Header.Message = "ok"

	if len(req.URL.Query()["key"]) > 0 {
		key := req.URL.Query()["key"][0]
		b, e := indexDBMap.Get(key)

		if e == nil {
			var v = Index_db_struct{
				Key:   key,
				Value: string(b),
			}
			util.JsonResponse(rw, v)
		} else {
			response.Header.Message = e.Error()
		}

	}
	util.JsonResponse(rw, response)
	return
}
