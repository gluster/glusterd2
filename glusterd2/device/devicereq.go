package device

import "github.com/pborman/uuid"

// AddDeviceReq structure
type AddDeviceReq struct {
	NodeID     uuid.UUID `json:"nodeid"`
	DeviceName []string  `json:"devicename"`
}
