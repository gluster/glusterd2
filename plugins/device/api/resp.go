package api

import (
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
)

// Info represents structure in which devices are to be store in Peer Metadata
type Info struct {
	Device        string    `json:"device"`
	State         string    `json:"state"`
	AvailableSize uint64    `json:"available-size"`
	ExtentSize    uint64    `json:"extent-size"`
	Used          bool      `json:"used"`
	PeerID        uuid.UUID `json:"peer-id"`
}

// AddDeviceResp is the success response sent to a AddDeviceReq request
type AddDeviceResp api.Peer

// ListDeviceResp is the success response sent to a ListDevice request
type ListDeviceResp []Info
