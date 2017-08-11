package api

// HTTPError represents HTTP error returned by glusterd2
type HTTPError struct {
	Error string `json:"Error"`
}
