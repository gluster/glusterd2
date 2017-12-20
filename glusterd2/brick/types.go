package brick

import (
	"github.com/pborman/uuid"
)

// Type is the type of Brick
//go:generate stringer -type=Type
type Type uint16

const (
	// Brick represents default type of brick
	Brick Type = iota
	// Arbiter represents Arbiter brick type
	Arbiter
)

// Brickinfo is the static information about the brick
type Brickinfo struct {
	ID         uuid.UUID
	Hostname   string
	NodeID     uuid.UUID
	Path       string
	VolumeName string
	VolumeID   uuid.UUID
	Type       Type
}

func (b *Brickinfo) String() string {
	return b.NodeID.String() + ":" + b.Path
}

// StringMap returns a map[string]string representation of the Brickinfo
func (b *Brickinfo) StringMap() map[string]string {
	m := make(map[string]string)

	m["brick.id"] = b.ID.String()
	m["brick.hostname"] = b.Hostname
	m["brick.nodeid"] = b.NodeID.String()
	m["brick.path"] = b.Path
	m["brick.volumename"] = b.VolumeName
	m["brick.volumeid"] = b.VolumeID.String()

	return m
}
