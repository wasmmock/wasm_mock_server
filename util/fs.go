package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

func WasmReplace(targets []string, slave []byte) error {
	for _, t := range targets {
		err := ioutil.WriteFile("wasm_modules/"+t+".wasm", slave, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
func WasmForfeit(wasmFileName string) {
	if _, err := os.Stat("wasm_modules/" + wasmFileName + ".wasm"); os.IsExist(err) {
		// path/to/whatever does not exist
		e := os.Rename("wasm_modules/"+wasmFileName+".wasm", "wasm_modules/"+wasmFileName+".wasm.bak")
		if e != nil {
			log.Fatal(e)
		}
	}

}
func WasmUnforfeit(wasmFileName string) {
	if _, err := os.Stat("wasm_modules/" + wasmFileName + ".wasm.bak"); os.IsExist(err) {
		e := os.Rename("wasm_modules/"+wasmFileName+".wasm.bak", "wasm_modules/"+wasmFileName+".wasm")
		if e != nil {
			log.Fatal(e)
		}
	}
}
func SaveReport(uId string, report Report) error {
	dBtye, err := json.Marshal(report)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("report/"+uId+".json", dBtye, 0644)
	if err != nil {
		return err
	}

	return err
}
