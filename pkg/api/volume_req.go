package api

// VolCreateReq represents a Volume Create Request
type VolCreateReq struct {
	Name      string            `json:"name"`
	Transport string            `json:"transport,omitempty"`
	Replica   int               `json:"replica,omitempty"`
	Bricks    []string          `json:"bricks"`
	Options   map[string]string `json:"options,omitempty"`
	Force     bool              `json:"force,omitempty"`
}

// VolOptionReq represents an incoming request to set volume options
type VolOptionReq struct {
	Options map[string]string `json:"options"`
}

// VolExpandReq represents a request to expand the volume by adding more bricks
type VolExpandReq struct {
	ReplicaCount int      `json:"replica,omitempty"`
	Bricks       []string `json:"bricks"`
}
