package api

// SnapInfo contains static information about the volume.
// Clients should NOT use this struct directly.
type SnapInfo struct {
	VolInfo       VolumeInfo `json:"snapinfo"`
	ParentVolName string     `json:"parentname"`
	Description   string     `json:"description"`
}

// SnapCreateResp is the response sent for a volume create request.
type SnapCreateResp SnapInfo

// SnapGetResp is the response sent for a volume get request.
type SnapGetResp SnapInfo

// SnapListResp is the response sent for a volume list request.
type SnapListResp []SnapGetResp
