package api

// SnapCreateReq represents a Snapshot Create Request
type SnapCreateReq struct {
	VolName     string `json:"volname"`
	SnapName    string `json:"snapname"`
	TimeStamp   bool   `json:"timestamp,omitempty"`
	Description string `json:"description,omitempty"`
	Force       bool   `json:"force,omitempty"`
}

//SnapActivateReq represents a request to activate a snapshot
type SnapActivateReq struct {
	Force bool `json:"force,omitempty"`
}

//SnapCloneReq represents a request to clone a snapshot
type SnapCloneReq struct {
	CloneName string `json:"clonename"`
}
