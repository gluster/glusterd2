package restclient

import (
	"fmt"
)

// UnexpectedStatusError is custom error when expected
// status code does not match with return status code
type UnexpectedStatusError struct {
	msg      string
	expected int
	actual   int
	resp     string
}

func (e *UnexpectedStatusError) Error() string {
	return fmt.Sprintf("Request failed with HTTP Status code %d\n%s", e.actual, e.resp)
}
