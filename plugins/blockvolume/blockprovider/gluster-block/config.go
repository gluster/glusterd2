package glusterblock

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

// ClientConfig holds various config information needed to create a gluster-block rest client
type ClientConfig struct {
	HostAddress string
	User        string
	Secret      string
	CaCertFile  string
	Insecure    bool
}

// ApplyFromConfig sets the ClientConfig options from various config sources
func (c *ClientConfig) ApplyFromConfig(conf *viper.Viper) {
	c.CaCertFile = conf.GetString("gluster-block-cacert")
	c.HostAddress = conf.GetString("gluster-block-hostaddr")
	c.User = conf.GetString("gluster-block-user")
	c.Secret = conf.GetString("gluster-block-secret")
	c.Insecure = conf.GetBool("gluster-block-insecure")
}

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
