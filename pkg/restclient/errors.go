package restclient

import (
	"fmt"
)

// APIUnexpectedStatusCodeError is custom error when expected
// status code does not match with return status code
type APIUnexpectedStatusCodeError struct {
	msg      string
	expected int
	actual   int
	resp     string
}

func raiseAPIUnexpectedStatusCodeError(expected int, actual int, resp string) error {
	return &APIUnexpectedStatusCodeError{"Unexpected Status Code", expected, actual, resp}
}

func (e *APIUnexpectedStatusCodeError) Error() string {
	msg := fmt.Sprintf("Status Code: %d Error: %s", e.actual, e.resp)
	return msg
}
