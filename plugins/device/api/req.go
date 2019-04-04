package api

const (
	// DeviceEnabled represents enabled
	DeviceEnabled = "enabled"

	// DeviceDisabled represents disabled
	DeviceDisabled = "disabled"
)

// AddDeviceReq structure
type AddDeviceReq struct {
	Device          string `json:"device"`
	ProvisionerType string `json:"provisioner"`
}

// EditDeviceReq structure
type EditDeviceReq struct {
	State string `json:"state"`
}
