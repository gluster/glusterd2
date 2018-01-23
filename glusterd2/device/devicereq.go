package device

import "github.com/pborman/uuid"

// Adding Device Request structure for Gd2
type AddDeviceReq struct {
	NodeID     uuid.UUID `json:"nodeid"`
	DeviceName []string  `json:"devicename"`
}
