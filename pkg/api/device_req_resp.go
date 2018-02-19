package api

import "github.com/pborman/uuid"

const (
	// DeviceEnabled represents enabled
	DeviceEnabled = "Enabled"

	// DeviceDisabled represents disabled
	DeviceDisabled = "Disabled"

	// DeviceFailed represents failed
	DeviceFailed = "Failed"
)

// AddDeviceReq structure
type AddDeviceReq struct {
	Devices  []string  `json:"devices"`
}

// Device is the added device info
type Device struct {
	PeerID uuid.UUID `json:"peer-id"`
	Detail []Info    `json:"device-details"`
}

// Info structure is the info of each device
type Info struct {
	Name  string `json:"name"`
	State string `json:"state"`
}
