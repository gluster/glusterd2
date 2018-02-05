package device

import "github.com/pborman/uuid"

// AddDeviceReq structure
type AddDeviceReq struct {
	PeerID uuid.UUID `json:"peerid"`
	Names  []string  `json:"names"`
}
