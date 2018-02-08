package api

// GlobalOptionReq represents an incoming request to set cluster level options
type GlobalOptionReq struct {
	Options map[string]string `json:"options"`
}
