package api

// HTTPError contains an error code and corresponding text which briefly
// describes the error in short.
type HTTPError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// ErrorResp is an error response which may contain one or more error responses
type ErrorResp struct {
	Errors []HTTPError `json:"errors"`
}

// ErrorCode represents API Error code Type
type ErrorCode uint16

const (
	// ErrCodeGeneric represents generic error code for API responses
	ErrCodeGeneric ErrorCode = iota + 1
	// ErrTxnStepFailed represents failure of a txn step
	ErrTxnStepFailed
)

// ErrorCodeMap maps error code to it's textual message
var ErrorCodeMap = map[ErrorCode]string{
	ErrCodeGeneric:   "generic error",
	ErrTxnStepFailed: "a txn step failed",
}

// ErrorResponse is an interface that types can implement on custom errors.
type ErrorResponse interface {
	error
	// Response returns the error response to be returned to client.
	Response() ErrorResp
	// Status returns the HTTP status code to be returned to client.
	Status() int
}
