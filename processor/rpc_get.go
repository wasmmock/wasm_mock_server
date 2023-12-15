package processor

import (
	"net/http"
	"strings"
)

func CommandsGet() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		data := strings.Join(registeredCommands, ",")
		buf := []byte(data)
		rw.Write(buf)
	}
}
