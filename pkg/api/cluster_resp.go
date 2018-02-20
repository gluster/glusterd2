package api

// GlobalOptionsGetResp contains details for global options
type GlobalOptionsGetResp struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	DefaultValue string `json:"default"`
	Modified     bool   `json:"modified"`
}
