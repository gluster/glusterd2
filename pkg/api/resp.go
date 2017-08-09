package api

import (
	"github.com/pborman/uuid"
)

// Peer reperesents a GlusterD
type Peer struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Addresses []string  `json:"addresses"`
}

// VolState is the current status of a volume
type VolState uint16

// VolType is the status of the volume
type VolType uint16

// VolAuth represents username and password used by trusted/internal clients
type VolAuth struct {
	Username string
	Password string
}

// Brickinfo is the static information about the brick
type Brickinfo struct {
	Hostname   string
	NodeID     uuid.UUID
	Path       string
	VolumeName string
	VolumeID   uuid.UUID
}

// Volinfo repesents a volume
type Volinfo struct {
	ID           uuid.UUID
	Name         string
	Type         VolType
	Transport    string
	DistCount    int
	ReplicaCount int
	Options      map[string]string
	Status       VolState
	Checksum     uint64
	Version      uint64
	Bricks       []Brickinfo
	Auth         VolAuth // TODO: should not be returned to client
}
