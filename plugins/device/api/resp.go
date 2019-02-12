package api

import (
	"strings"

	"github.com/pborman/uuid"
)

// Info represents structure in which devices are to be store in Peer Metadata
type Info struct {
	Device          string    `json:"device"`
	ProvisionerType string    `json:"provisioner"`
	State           string    `json:"state"`
	AvailableSize   uint64    `json:"free-size"`
	TotalSize       uint64    `json:"total-size"`
	UsedSize        uint64    `json:"used-size"`
	ExtentSize      uint64    `json:"extent-size"`
	Used            bool      `json:"device-used"`
	PeerID          uuid.UUID `json:"peer-id"`
}

// VgName returns name for LVM Vg
func (info *Info) VgName() string {
	return "gluster" + strings.Replace(info.Device, "/", "-", -1)
}

// AddDeviceResp is the success response sent to a AddDeviceReq request
type AddDeviceResp Info

// ListDeviceResp is the success response sent to a ListDevice request
type ListDeviceResp []Info
