package processor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/wasmmock/wasm_mock_server/internal/box"
	"github.com/wasmmock/wasm_mock_server/logger"
	"github.com/wasmmock/wasm_mock_server/model"
	"github.com/wasmmock/wasm_mock_server/util"
	"github.com/gorilla/mux"
)

var tmpl template.Template

func Report(rw http.ResponseWriter, req *http.Request) {
	reportPath := mux.Vars(req)["uid"]
	// Create a template.
	// Parse template file
	fmt.Println("reportPath!!", reportPath)
	jsonFile, err := os.Open("report/" + reportPath + ".json")
	if err != nil {
		return
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	stringValue := strings.ReplaceAll(string(byteValue), "'", "&apos;")
	if err != nil {
		return
	}
	var report util.Report
	json.Unmarshal([]byte(stringValue), &report)
	tmpl.Execute(rw, report)
}
func ReportData(rw http.ResponseWriter, req *http.Request) {
	reportPath := mux.Vars(req)["uid"]
	jsonFile, err := os.Open("report/" + reportPath + ".json")
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	response := new(model.CallApiResponse)
	if err != nil {
		response.Header = model.Header{
			Message:  err.Error(),
			ReportId: reportPath,
		}
		err = json.Unmarshal(byteValue, &response.ResponseBody)
		if err != nil {
			logger.Errorf("Failed to marshal response | resp=%s, err=%s", string(byteValue), err.Error())
		}
		util.JsonResponse(rw, response)
		return
	}
	err = json.Unmarshal(byteValue, &response.ResponseBody)
	if err != nil {
		logger.Errorf("Failed to marshal response | resp=%s, err=%s", string(byteValue), err.Error())
	}
	util.JsonResponse(rw, response)
	return
}
func TemplateInit() {
	//tmpl = *template.Must(template.ParseGlob("tmpl/*html"))
	code := box.Get("/report.html")
	s := string(code)
	tmplz, err := template.New("foo").Parse(s)
	if err == nil {
		tmpl = *tmplz
	}
}
