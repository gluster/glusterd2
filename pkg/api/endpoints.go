package api

// Endpoint represents an HTTP API endpoint.
type Endpoint struct {
	Name    string   `json:"name"`
	Methods []string `json:"methods"`
	Path    string   `json:"path"`
}

// ListEndpointsResp is the response sent to client for a list endpoints request.
type ListEndpointsResp []Endpoint
