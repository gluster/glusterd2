package api

// HTTPError contains an error code and corresponding text which briefly
// describes the error in short.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
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
)

// ErrorCodeMap maps error code to it's textual message
var ErrorCodeMap = map[ErrorCode]string{
	ErrCodeGeneric: "generic error",
}
