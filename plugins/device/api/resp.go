package api

import (
	"github.com/gluster/glusterd2/pkg/api"
)

// Info represents structure in which devices are to be store in Peer Metadata
type Info struct {
	Name          string `json:"name"`
	State         string `json:"state"`
	VgName        string `json:"vg-name"`
	AvailableSize uint64 `json:"available-size"`
	ExtentSize    uint64 `json:"extent-size"`
	Used          bool   `json:"used"`
	PeerID        string `json:"peer-id"`
}

// AddDeviceResp is the success response sent to a AddDeviceReq request
type AddDeviceResp api.Peer

// ListDeviceResp is the success response sent to a ListDevice request
type ListDeviceResp []Info
