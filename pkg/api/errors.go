package api

// HTTPError represents HTTP error returned by glusterd2
type HTTPError struct {
	Error string `json:"Error"`
}

// ErrorCode represents API Error code Type
type ErrorCode uint16

const (
	// ErrCodeDefault represents default error code for API responses
	ErrCodeDefault ErrorCode = iota + 1
)
