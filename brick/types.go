package brick

import (
	"github.com/pborman/uuid"
)

// Brickinfo is the static information about the brick
type Brickinfo struct {
	Hostname   string
	NodeID     uuid.UUID
	Path       string
	VolumeName string
	VolumeID   uuid.UUID
}

// Brickstatus represents real-time status of the brick and contains dynamic
// information about the brick
type Brickstatus struct {
	BInfo  Brickinfo
	Online bool
	Pid    int
	Port   int
	// TODO: Add other fields like filesystem type, statvfs output etc.
}
