package blockvolume

import (
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
	"github.com/spf13/viper"
)

// VolumeType represents a volume type
type VolumeType string

const (
	// Replica represents a replica volume type
	Replica VolumeType = "Replica"
)

// HostingVolumeOptions holds various information which will be used in creating hosting volume
type HostingVolumeOptions struct {
	Size         int64
	Type         VolumeType
	ReplicaCount int
	AutoCreate   bool
}

// ApplyFromConfig sets HostingVolumeOptions member values from given config source
func (h *HostingVolumeOptions) ApplyFromConfig(conf *viper.Viper) {
	h.Size = conf.GetInt64("block-hosting-volume-size")
	h.Type = VolumeType(conf.GetString("block-hosting-volume-type"))
	h.ReplicaCount = conf.GetInt("block-hosting-volume-replica-count")
	h.AutoCreate = conf.GetBool("auto-create-block-hosting-volumes")
}

// PrepareVolumeCreateReq will create a request body to be use for creating a gluster volume
func (h *HostingVolumeOptions) PrepareVolumeCreateReq() *api.VolCreateReq {
	name := "block_hosting_volume_" + uuid.NewRandom().String()

	req := &api.VolCreateReq{
		Name:         name,
		Transport:    "tcp",
		Size:         uint64(h.Size),
		ReplicaCount: h.ReplicaCount,
		SubvolType:   string(h.Type),
	}

	return req
}
