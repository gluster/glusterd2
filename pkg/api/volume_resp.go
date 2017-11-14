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
}

// BrickStatus contains the runtime information about the brick.
// Clients should NOT use this struct directly.
type BrickStatus struct {
	Info   BrickInfo `json:"info"`
	Online bool      `json:"online"`
	Pid    int       `json:"pid"`
	Port   int       `json:"port"`
}

// VolState is the current state of the volume.
type VolState uint16

// VolType is the type of volume.
type VolType uint16

// VolumeInfo contains static information about the volume.
// Clients should NOT use this struct directly.
type VolumeInfo struct {
	ID           uuid.UUID         `json:"id"`
	Name         string            `json:"name"`
	Type         VolType           `json:"type"`
	Transport    string            `json:"transport"`
	DistCount    int               `json:"distribute-count"`
	ReplicaCount int               `json:"replica-count"`
	Options      map[string]string `json:"options"`
	State        VolState          `json:"state"`
	Bricks       []BrickInfo       `json:"bricks"`
}

// VolumeStatusResp response contains the statuses of all bricks of the volume.
type VolumeStatusResp struct {
	Bricks []BrickStatus `json:"bricks"`
	// TODO: Add clients connected, capacity, free size etc.
}

// VolumeCreateResp is the response sent for a volume create request.
type VolumeCreateResp VolumeInfo

// VolumeGetResp is the response sent for a volume get request.
type VolumeGetResp VolumeInfo

// VolumeExpandResp is the response sent for a volume expand request.
type VolumeExpandResp VolumeInfo

// VolumeListResp is the response sent for a volume list request.
type VolumeListResp []VolumeGetResp
