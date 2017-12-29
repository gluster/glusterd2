package heketi

// AddDeviceReq represents REST API request to add device to heketi managed device list
type AddDeviceReq struct {
	NodeID     string `json:"nodeid"`
	DeviceName string `json:"devicename"`
}
