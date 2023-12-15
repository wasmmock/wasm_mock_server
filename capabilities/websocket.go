package capabilities

type WsProtocol struct {
	Fn         string `json:"fn"`
	Payload    string `json:"payload"`
	CallerAddr string `json:"caller_addr"`
}
type WsCallProtocol struct {
	Loop    string `json:"loop"`
	Fn      string `json:"fn"`
	Payload string `json:"payload"`
	Index   int64  `json:"index"`
	Binding string `json:"binding"`
	Message string `json:"message"`
	Targets string `json:"targets"`
}
