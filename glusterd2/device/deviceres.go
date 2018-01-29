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
	NodeID     uuid.UUID `json:"nodeid"`
	DeviceName []string  `json:"devicename"`
	State      string    `json:"devicestate"`
}
