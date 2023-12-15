package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

func JsonResponse(rw http.ResponseWriter, reply interface{}) {
	js, err := json.Marshal(reply)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err = rw.Write(js)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
func JsonResponseWS(rw *websocket.Conn, reply interface{}) {
	err := rw.WriteJSON(reply)
	if err != nil {
		fmt.Println("err writejson")
		return
	}
}
func GetPath(url *url.URL) (result string) {
	for result = strings.TrimSuffix(url.EscapedPath(), "/"); strings.HasSuffix(result, "/"); result = strings.TrimSuffix(result, "/") {
	}
	return
}
