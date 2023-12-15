package model

const (
	RESULT_SUCCESS      = 0
	RESULT_ERROR_SYSTEM = 1
	RESULT_ERROR_PARAM  = 2
)

type Header struct {
	Result   uint32
	Message  string
	auth     string
	ReportId string
}

type CallApiRequest struct {
	Command       string
	RequestBody   map[string]interface{} `json:"request_body"`
	RequestProto  string                 `json:"request_proto"`
	ResponseProto string                 `json:"response_proto"`
}

type CallApiResponse struct {
	Header       Header
	ResponseBody map[string]interface{} `json:"response_body"`
}
