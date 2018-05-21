package device

const (
	// DeviceEnabled represents enabled
	DeviceEnabled = "Enabled"

	// DeviceDisabled represents disabled
	DeviceDisabled = "Disabled"
)

// AddDeviceReq structure
type AddDeviceReq struct {
	Device string `json:"device"`
}
