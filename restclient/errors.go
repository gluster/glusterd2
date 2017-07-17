package restclient

type APIUnsupportedError struct {
	msg string
}

type APIUnexpectedStatusCodeError struct {
	msg      string
	expected int
	actual   int
	resp     string
}

func NewAPIUnsupportedError() error {
	return &APIUnsupportedError{"Unsupported API"}
}

func (e *APIUnsupportedError) Error() string {
	return e.msg
}

func NewAPIUnexpectedStatusCodeError(expected int, actual int, resp string) error {
	return &APIUnexpectedStatusCodeError{"Unexpected Status Code", expected, actual, resp}
}

func (e *APIUnexpectedStatusCodeError) Error() string {
	msg := fmt.Sprintf("%s Expected: %d Actual: %d (%s)", e.msg, e.expected, e.actual, e.resp)
	return msg
}

