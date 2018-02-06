package device

import "github.com/pborman/uuid"

// AddDeviceReq structure
type AddDeviceReq struct {
	PeerID uuid.UUID `json:"peer-id"`
	Names  []string  `json:"names"`
}
