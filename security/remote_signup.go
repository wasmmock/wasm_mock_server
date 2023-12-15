package Security

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Request struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func Signup() {
	requestBody := Request{}
	n, _ := json.Marshal(requestBody)
	resp, err := http.Post(RemoteVerificationState.Addr+"/remote_signup", "application/json", bytes.NewBuffer(n))
	_ = resp
	_ = err
}
