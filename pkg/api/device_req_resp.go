package api

import "github.com/pborman/uuid"

// AddDeviceReq structure
type AddDeviceReq struct {
        PeerID uuid.UUID `json:"peer-id"`
        Names  []string  `json:"names"`
}

type Info struct {
        PeerID uuid.UUID `json:"peer-id"`
        Names  []string  `json:"names"`
        State  string    `json:"state"`
}
