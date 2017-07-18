package restclient

import (
	"fmt"
)

// APIUnsupportedError is custom error struct for API 404 errors
type APIUnsupportedError struct {
	msg string
}

// APIUnexpectedStatusCodeError is custom error when expected
// status code does not match with return status code
type APIUnexpectedStatusCodeError struct {
	msg      string
	expected int
	actual   int
	resp     string
}

func raiseAPIUnsupportedError() error {
	return &APIUnsupportedError{"Unsupported API"}
}

func (e *APIUnsupportedError) Error() string {
	return e.msg
}

func raiseAPIUnexpectedStatusCodeError(expected int, actual int, resp string) error {
	return &APIUnexpectedStatusCodeError{"Unexpected Status Code", expected, actual, resp}
}

func (e *APIUnexpectedStatusCodeError) Error() string {
	msg := fmt.Sprintf("%s Expected: %d Actual: %d (%s)", e.msg, e.expected, e.actual, e.resp)
	return msg
}
