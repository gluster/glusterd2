package heketi

import "github.com/pborman/uuid"

const (
	// HeketiDeviceEnabled represents enabled
	HeketiDeviceEnabled = "Enabled"

	// HeketiDeviceFrozen represents frozen
	HeketiDeviceFrozen = "Frozen"

	// HeketiDeviceEvacuated represents evacuated
	HeketiDeviceEvacuated = "Evacuated"
)

// DeviceInfo is the added device info
type DeviceInfo struct {
	NodeID     uuid.UUID `json:"nodeid"`
	DeviceName string    `json:"devicename"`
	State      string    `json:"devicestate"`
}
