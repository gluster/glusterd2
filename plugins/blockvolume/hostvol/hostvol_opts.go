package hostvol

import (
	"errors"
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/options"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/size"

	blkapi "github.com/gluster/glusterd2/plugins/blockvolume/api"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	defaultHostVolsize         = size.GiB * 5
	defaultHostVolType         = "Replicate"
	defaultHostVolReplicaCount = 3
	hostVolautoCreate          = true
	hostVolautoDelete          = true
)

// HostingVolumeOptions holds various information which will be used in creating hosting volume
type HostingVolumeOptions struct {
	Size         size.Size
	Type         string
	ReplicaCount int
	AutoCreate   bool
	AutoDelete   bool
	ThinArbPath  string
	ShardSize    uint64
}

func newHostingVolumeOptions() *HostingVolumeOptions {
	return &HostingVolumeOptions{
		Size:         defaultHostVolsize,
		Type:         defaultHostVolType,
		ReplicaCount: defaultHostVolReplicaCount,
		AutoCreate:   hostVolautoCreate,
		AutoDelete:   hostVolautoDelete,
	}
}

// PrepareVolumeCreateReq will create a request body to be use for creating a gluster volume
func (h *HostingVolumeOptions) PrepareVolumeCreateReq() (*api.VolCreateReq, error) {
	name := "block_hosting_volume_" + uuid.NewRandom().String()

	req := &api.VolCreateReq{
		Name:         name,
		Transport:    "tcp",
		Size:         uint64(h.Size),
		ReplicaCount: h.ReplicaCount,
		SubvolType:   h.Type,
		Force:        true,
		VolOptionReq: api.VolOptionReq{
			Options: map[string]string{},
		},
	}

	if h.ThinArbPath != "" {
		if h.ReplicaCount != 2 {
			err := errors.New("thin arbiter can only be enabled for replica count 2")
			log.WithError(err).Error("failed to prepare host vol create request")
			return nil, err
		}
		if err := volume.AddThinArbiter(req, h.ThinArbPath); err != nil {
			log.WithError(err).Error("failed to add thin arbiter options to host volume")
			return nil, err
		}
	}
	if h.ShardSize != 0 {
		volume.AddShard(req, h.ShardSize)
	}

	return req, nil
}

// SetFromReq will configure HostingVolumeOptions from the values sent in the block create request
func (h *HostingVolumeOptions) SetFromReq(hvi *blkapi.HostVolumeInfo) {
	if hvi.HostVolReplicaCnt != 0 {
		h.ReplicaCount = hvi.HostVolReplicaCnt
	}
	if hvi.HostVolThinArbPath != "" {
		h.ThinArbPath = hvi.HostVolThinArbPath
	}
	if hvi.HostVolShardSize != 0 {
		h.ShardSize = hvi.HostVolShardSize
	}
	if hvi.HostVolSize != 0 {
		h.Size = size.Size(hvi.HostVolSize)
	}
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

	autoDelete, err := options.GetClusterOption("auto-delete-block-hosting-volumes")
	if err == nil {
		if val, err := strconv.ParseBool(autoDelete); err == nil {
			h.AutoDelete = val
		}
	}

	h.ThinArbPath = ""
	h.ShardSize = 0
}
