package device

import (
	"github.com/gluster/glusterd2/pkg/api"
)

// Info represents structure in which devices are to be store in Peer MetaData
type Info struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

// AddDeviceResp is the success response sent to a AddDeviceReq request
type AddDeviceResp api.Peer
