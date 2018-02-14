package api

// Endpoint represents an HTTP API endpoint.
type Endpoint struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Method       string `json:"methods"`
	Path         string `json:"path"`
	RequestType  string `json:"request-type"`
	ResponseType string `json:"response-type"`
}

// ListEndpointsResp is the response sent to client for a list endpoints request.
type ListEndpointsResp []Endpoint
