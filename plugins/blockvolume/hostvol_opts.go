package blockvolume

import (
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/options"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/size"

	"github.com/pborman/uuid"
)

const (
	defaultHostVolsize         = size.GiB * 5
	defaultHostVolType         = "Replicate"
	defaultHostVolReplicaCount = 3
	hostVolautoCreate          = true
)

// HostingVolumeOptions holds various information which will be used in creating hosting volume
type HostingVolumeOptions struct {
	Size         size.Size
	Type         string
	ReplicaCount int
	AutoCreate   bool
}

func newHostingVolumeOptions() *HostingVolumeOptions {
	return &HostingVolumeOptions{
		Size:         defaultHostVolsize,
		Type:         defaultHostVolType,
		ReplicaCount: defaultHostVolReplicaCount,
		AutoCreate:   hostVolautoCreate,
	}
}

// PrepareVolumeCreateReq will create a request body to be use for creating a gluster volume
func (h *HostingVolumeOptions) PrepareVolumeCreateReq() *api.VolCreateReq {
	name := "block_hosting_volume_" + uuid.NewRandom().String()

	req := &api.VolCreateReq{
		Name:         name,
		Transport:    "tcp",
		Size:         uint64(h.Size),
		ReplicaCount: h.ReplicaCount,
		SubvolType:   h.Type,
	}

	return req
}

// SetFromClusterOptions will configure HostingVolumeOptions using cluster options
func (h *HostingVolumeOptions) SetFromClusterOptions() {
	volType, err := options.GetClusterOption("block-hosting-volume-type")
	if err == nil {
		h.Type = volType
	}

	volSize, err := options.GetClusterOption("block-hosting-volume-size")
	if err == nil {
		if hostVolSize, err := size.Parse(volSize); err == nil {
			h.Size = hostVolSize
		}
	}

	count, err := options.GetClusterOption("block-hosting-volume-replica-count")
	if err == nil {
		if replicaCount, err := strconv.Atoi(count); err == nil {
			h.ReplicaCount = replicaCount
		}
	}

	autoCreate, err := options.GetClusterOption("auto-create-block-hosting-volumes")
	if err == nil {
		if val, err := strconv.ParseBool(autoCreate); err == nil {
			h.AutoCreate = val
		}
	}
}
