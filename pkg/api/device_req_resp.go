package api

import "github.com/pborman/uuid"

const (
        // DeviceEnabled represents enabled
        DeviceEnabled = "Enabled"

        // DeviceFrozen represents frozen
        DeviceFrozen = "Frozen"

        // DeviceEvacuated represents evacuated
        DeviceEvacuated = "Evacuated"

	// DeviceFailed represents failed
	DeviceFailed = "Failed"
)


// AddDeviceReq structure
type AddDeviceReq struct {
	PeerID uuid.UUID `json:"peer-id"`
	Names  []string  `json:"names"`
}

// Info structure
type Info struct {
	PeerID uuid.UUID `json:"peer-id"`
	Names  []string  `json:"names"`
	State  string    `json:"state"`
}
