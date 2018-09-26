package api

import (
	"time"

	"github.com/pborman/uuid"
)

// SnapInfo contains static information about the snapshot.
// Clients should NOT use this struct directly.
type SnapInfo struct {
	VolInfo       VolumeInfo `json:"snapinfo"`
	ParentVolName string     `json:"parentname"`
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"created-at"`
}

//SnapList contains snapshots information of a volume.
type SnapList struct {
	ParentName string     `json:"parentname"`
	SnapList   []SnapInfo `json:"snaps"`
}

//LvsData gives the information provided in lvs command
type LvsData struct {
	VgName         string  `json:"vgname"`
	DataPercentage float32 `json:"datapercentage"`
	LvSize         string  `json:"lvsize"`
	PoolLV         string  `json:"pool-lv"`
}

//SnapBrickStatus contains information about a snap brick
type SnapBrickStatus struct {
	Brick  BrickStatus `json:"brick"`
	LvData LvsData     `json:"lvs-data"`
}

//SnapStatusResp contains snapshot status
type SnapStatusResp struct {
	ParentName  string            `json:"parentname"`
	SnapName    string            `json:"snaps"`
	ID          uuid.UUID         `json:"id"`
	BrickStatus []SnapBrickStatus `json:"snapbrickstatus"`
}

// SnapCreateResp is the response sent for a snapshot create request.
type SnapCreateResp SnapInfo

// SnapGetResp is the response sent for a snapshot get request.
type SnapGetResp SnapInfo

// SnapListResp is the response sent for a snapsht list request.
type SnapListResp []SnapList

// SnapshotActivateResp is the response sent for a snapshot activate request.
type SnapshotActivateResp SnapInfo

// SnapshotDeactivateResp is the response sent for a snapshot deactivate request.
type SnapshotDeactivateResp SnapInfo

// SnapshotCloneResp is the response sent for a snapshot clone request.
// Snapshot clone will create a regular volume
type SnapshotCloneResp VolumeInfo
