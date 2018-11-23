package api

// ClusterOptionReq represents an incoming request to set cluster level options
type ClusterOptionReq struct {
	Options map[string]string `json:"options"`
}

// ClusterOptionsResp contains details for global options
type ClusterOptionsResp struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	DefaultValue string `json:"default"`
	Modified     bool   `json:"modified"`
}
