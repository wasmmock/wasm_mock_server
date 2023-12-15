/*
Error handling module
*/
package myerror

//Code is ErrorCode Enum type
type Code uint

//Create ErrorCode enum
const (
	Nil Code = iota
	AddRecordFail
	DeleteRecordFail
	UpdateRecordFail
	NoRecords
	NoImplementation
	RowScanFailure
	DBQueryFailure
	DBError
	ProtobufMarshalError
	MultiRequestError
	RequestTypeUndefined
	MissingFields
	ResponseTimeout
	HttpMethodUnMatched
)

//Error method is implemented for ErrorCode that will return the error message
func (o Code) Error() string {
	switch o {
	case AddRecordFail:
		return "Add Record fail"
	case DeleteRecordFail:
		return "Delete Record fail"
	case UpdateRecordFail:
		return "Update Record fail"
	case NoRecords:
		return "No record found"
	case NoImplementation:
		return "No such implementation"
	case RowScanFailure:
		return "Database row scan fail"
	case DBQueryFailure:
		return "Database failure"
	case DBError:
		return "Database Error"
	case ProtobufMarshalError:
		return "Wrong Protobuf format"
	case MultiRequestError:
		return "Multi Requests detected, unable to decide what to execute"
	case RequestTypeUndefined:
		return "Request param not defined"
	case MissingFields:
		return "Missing parameters, please check the api documentation"
	case ResponseTimeout:
		return "Response time out after 5 seconds"
	case HttpMethodUnMatched:
		return "HttpMethodUnMatched"
	default:
		return ""
	}
}
