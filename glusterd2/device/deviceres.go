package device

import "github.com/pborman/uuid"

const (
	// DeviceEnabled represents enabled
	DeviceEnabled = "Enabled"

	// DeviceFrozen represents frozen
	DeviceFrozen = "Frozen"

	// DeviceEvacuated represents evacuated
	DeviceEvacuated = "Evacuated"
)

// Info is the added device info
type Info struct {
	PeerID uuid.UUID `json:"peerid"`
	Names  []string  `json:"names"`
	State  string    `json:"state"`
}
