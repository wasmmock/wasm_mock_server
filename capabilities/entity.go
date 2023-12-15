package capabilities

import "time"

type Entity struct {
	Id          string `json:"id"`
	Command     string `json:"command"`
	ResponseLen int64  `json:"response_len"`
}

type EntityWhole struct {
	Id          string
	Time        time.Time
	Command     string
	ResponseLen int64
	ReportType  string
	Payload     []byte
}
