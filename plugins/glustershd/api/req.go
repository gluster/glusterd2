package api

// SplitBrainReq represents details needed to resolve split brain
type SplitBrainReq struct {
	FileName  string `json:"filename,omitempty"`
	HostName  string `json:"hostname,omitempty"`
	BrickName string `json:"brickname,omitempty"`
}
