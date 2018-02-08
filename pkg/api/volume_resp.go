package api

import "github.com/pborman/uuid"

// BrickInfo contains the static information about the brick.
// Clients should NOT use this struct directly.
type BrickInfo struct {
	ID         uuid.UUID `json:"id"`
	Path       string    `json:"path"`
	VolumeID   uuid.UUID `json:"volume-id"`
	VolumeName string    `json:"volume-name"`
	NodeID     uuid.UUID `json:"node-id"`
	Hostname   string    `json:"host"`
	Type       BrickType `json:"type"`
}

// Subvol contains static information about sub volume
type Subvol struct {
	Name         string      `json:"name"`
	Type         SubvolType  `json:"type"`
	Bricks       []BrickInfo `json:"bricks"`
	Subvols      []Subvol    `json:"subvols,omitempty"`
	ReplicaCount int         `json:"replica-count"`
	ArbiterCount int         `json:"arbiter-count"`
}

// SizeInfo represents sizing information.
// Clients should NOT use this struct directly.
type SizeInfo struct {
	Capacity uint64 `json:"capacity"`
	Used     uint64 `json:"used"`
	Free     uint64 `json:"free"`
}

// BrickStatus contains the runtime information about the brick.
// Clients should NOT use this struct directly.
type BrickStatus struct {
	Info      BrickInfo `json:"info"`
	Online    bool      `json:"online"`
	Pid       int       `json:"pid"`
	Port      int       `json:"port"`
	FS        string    `json:"fs-type"`
	MountOpts string    `json:"mount-opts"`
	Device    string    `json:"device"`
	Size      SizeInfo  `json:"size"`
}

// BricksStatusResp contains statuses of bricks belonging to one
// volume.
type BricksStatusResp []BrickStatus

// VolumeInfo contains static information about the volume.
// Clients should NOT use this struct directly.
type VolumeInfo struct {
	ID           uuid.UUID         `json:"id"`
	Name         string            `json:"name"`
	Type         VolType           `json:"type"`
	Transport    string            `json:"transport"`
	DistCount    int               `json:"distribute-count"`
	ReplicaCount int               `json:"replica-count"`
	ArbiterCount int               `json:"arbiter-count"`
	Options      map[string]string `json:"options"`
	State        VolState          `json:"state"`
	Subvols      []Subvol          `json:"subvols"`
}

// VolumeStatusResp response contains the statuses of all bricks of the volume.
type VolumeStatusResp struct {
	Info   VolumeInfo `json:"info"`
	Online bool       `json:"online"`
	Size   SizeInfo   `json:"size"`
}

// VolumeCreateResp is the response sent for a volume create request.
type VolumeCreateResp VolumeInfo

// VolumeGetResp is the response sent for a volume get request.
type VolumeGetResp VolumeInfo

// VolumeExpandResp is the response sent for a volume expand request.
type VolumeExpandResp VolumeInfo

// VolumeListResp is the response sent for a volume list request.
type VolumeListResp []VolumeGetResp

// OptionGroupListResp is the response sent for a group list request.
type OptionGroupListResp []OptionGroup
