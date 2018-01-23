package api

// SnapCreateReq represents a Snapshot Create Request
type SnapCreateReq struct {
	VolName     string `json:"volname"`
	SnapName    string `json:"snapname"`
	TimeStamp   bool   `json:"timestamp,omitempty"`
	Description string `json:description,omitempty"`
	Force       bool   `json:"force,omitempty"`
}
