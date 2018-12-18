package glusterblock

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/size"
	"github.com/gluster/glusterd2/plugins/blockvolume/blockprovider"
	"github.com/gluster/glusterd2/plugins/blockvolume/utils"

	"github.com/gluster/gluster-block-restapi/client"
	"github.com/gluster/gluster-block-restapi/pkg/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const providerName = "gluster-block"

func init() {
	log.WithField("name", providerName).Infof("Registering block provider")
	blockprovider.RegisterBlockProvider(providerName, newGlusterBlock)
}

// GlusterBlock implements block Provider interface. It represents a gluster-block
type GlusterBlock struct {
	client     client.GlusterBlockClient
	ClientConf *ClientConfig
}

func newGlusterBlock() (blockprovider.Provider, error) {
	var (
		gb         = &GlusterBlock{}
		clientConf = &ClientConfig{}
		opts       = []client.OptFuncs{}
	)

	clientConf.ApplyFromConfig(viper.GetViper())
	gb.ClientConf = clientConf

	opts = append(opts,
		client.WithAuth(clientConf.User, clientConf.Secret),
		client.WithTLSConfig(&client.TLSOptions{CaCertFile: clientConf.CaCertFile, InsecureSkipVerify: clientConf.Insecure}),
	)

	gbClient, err := client.NewClientWithOpts(clientConf.HostAddress, opts...)
	if err != nil {
		return nil, err
	}
	gb.client = gbClient

	return gb, nil
}

// CreateBlockVolume will create a gluster block volume with given name and size having `hostVolume` as hosting volume
func (g *GlusterBlock) CreateBlockVolume(name string, size int64, hosts []string, hostVolume string, options ...blockprovider.BlockVolOption) (blockprovider.BlockVolume, error) {
	var (
		blockVolOpts = &blockprovider.BlockVolumeOptions{}
		clusterLocks = transaction.Locks{}
	)

	blockVolOpts.ApplyOpts(options...)

	req := &api.BlockVolumeCreateReq{
		HaCount:            blockVolOpts.Ha,
		AuthEnabled:        blockVolOpts.Auth,
		FullPrealloc:       blockVolOpts.FullPrealloc,
		Size:               uint64(size),
		Storage:            blockVolOpts.Storage,
		RingBufferSizeInMB: blockVolOpts.RingBufferSizeInMB,
		Hosts:              hosts,
	}

	volInfo, err := volume.GetVolume(hostVolume)
	if err != nil {
		return nil, fmt.Errorf("error in getting host vol details: %s", err)
	}

	if err := clusterLocks.Lock(volInfo.Name); err != nil {
		log.WithError(err).Error("error in acquiring cluster lock")
		return nil, err
	}

	defer clusterLocks.UnLock(context.Background())

	availableSizeInBytes, err := strconv.Atoi(volInfo.Metadata["block-hosting-available-size"])
	if err != nil {
		return nil, err
	}

	if int64(availableSizeInBytes) < size {
		return nil, fmt.Errorf("available size is less than requested size,request size: %d, available size: %d", size, availableSizeInBytes)
	}

	resp, err := g.client.CreateBlockVolume(hostVolume, name, req)
	if err != nil {
		return nil, err
	}

	volInfo.Metadata["block-hosting-available-size"] = fmt.Sprintf("%d", int64(availableSizeInBytes)-size)

	if err := volume.AddOrUpdateVolume(volInfo); err != nil {
		log.WithError(err).Error("failed in updating volume info to store")
	}

	return &BlockVolume{
		hostVolume: volInfo.Name,
		name:       name,
		hosts:      resp.Portals,
		iqn:        resp.IQN,
		username:   resp.Username,
		password:   resp.Password,
		size:       int64(size),
		ha:         blockVolOpts.Ha,
	}, nil
}

// DeleteBlockVolume deletes a gluster block volume of give name
func (g *GlusterBlock) DeleteBlockVolume(name string, options ...blockprovider.BlockVolOption) error {
	var (
		blockVolOpts = &blockprovider.BlockVolumeOptions{}
		hostVol      string
	)

	blockVolOpts.ApplyOpts(options...)

	blockVols := g.BlockVolumes()

	for _, blockVol := range blockVols {
		if blockVol.Name() == name {
			hostVol = blockVol.HostVolume()
			break
		}
	}

	if hostVol == "" {
		return errors.New("block volume not found")
	}

	blockInfo, err := g.client.BlockVolumeInfo(hostVol, name)
	if err != nil {
		return err
	}

	req := &api.BlockVolumeDeleteReq{
		UnlinkStorage: blockVolOpts.UnlinkStorage,
		Force:         blockVolOpts.ForceDelete,
	}

	if err := g.client.DeleteBlockVolume(hostVol, name, req); err != nil {
		return err
	}

	if err := utils.ResizeBlockHostingVolume(hostVol, blockInfo.Size); err != nil {
		log.WithError(err).Error("error in resizing the block hosting volume")
	}

	return nil
}

// GetBlockVolume gives info about a gluster block volume
func (g *GlusterBlock) GetBlockVolume(name string) (blockprovider.BlockVolume, error) {
	var (
		blockVolume           blockprovider.BlockVolume
		availableBlockVolumes = g.BlockVolumes()
	)

	for _, blockVol := range availableBlockVolumes {
		if blockVol.Name() == name {
			blockVolume = blockVol
			break
		}
	}

	if blockVolume == nil {
		return nil, errors.New("block volume not found")
	}

	blockInfo, err := g.client.BlockVolumeInfo(blockVolume.HostVolume(), blockVolume.Name())
	if err != nil {
		return nil, err
	}

	glusterBlockVol := &BlockVolume{
		name:       blockInfo.Name,
		hostVolume: blockInfo.Volume,
		password:   blockInfo.Password,
		hosts:      blockInfo.ExportedOn,
		gbID:       blockInfo.GBID,
		ha:         blockInfo.Ha,
	}

	if blockSize, err := size.Parse(blockInfo.Size); err == nil {
		glusterBlockVol.size = int64(blockSize)
	}

	return glusterBlockVol, nil
}

// BlockVolumes returns all available gluster block volume
func (g *GlusterBlock) BlockVolumes() []blockprovider.BlockVolume {
	var glusterBlockVolumes = []blockprovider.BlockVolume{}

	volumes, err := volume.GetVolumes(context.Background())
	if err != nil {
		return glusterBlockVolumes
	}

	volumes = volume.ApplyFilters(volumes, volume.BlockHosted)

	for _, vol := range volumes {
		blockList, err := g.client.ListBlockVolumes(vol.Name)
		if err != nil {
			continue
		}

		for _, block := range blockList.Blocks {
			glusterBlockVolumes = append(glusterBlockVolumes, &BlockVolume{name: block, hostVolume: vol.Name})
		}
	}

	return glusterBlockVolumes
}

// ProviderName returns name of block provider
func (g *GlusterBlock) ProviderName() string {
	return providerName
}

// BlockVolume implements blockprovider.BlockVolume interface.
// It holds information about a gluster-block volume
type BlockVolume struct {
	hosts      []string
	iqn        string
	username   string
	password   string
	hostVolume string
	name       string
	size       int64
	gbID       string
	ha         int
}

// HostAddresses returns host addresses of a gluster block vol
func (gv *BlockVolume) HostAddresses() []string { return gv.hosts }

// IQN returns IQN of a gluster block vol
func (gv *BlockVolume) IQN() string { return gv.iqn }

// Username returns username of a gluster-block vol.
func (gv *BlockVolume) Username() string { return gv.username }

// Password returns password for a gluster block vol
func (gv *BlockVolume) Password() string { return gv.password }

// HostVolume returns host vol name of gluster block
func (gv *BlockVolume) HostVolume() string { return gv.hostVolume }

// Name returns name of gluster block vol
func (gv *BlockVolume) Name() string { return gv.name }

// Size returns size of a gluster block vol in bytes
func (gv *BlockVolume) Size() uint64 { return uint64(gv.size) }

// ID returns Gluster Block ID
func (gv *BlockVolume) ID() string { return gv.gbID }

// HaCount returns high availability count
func (gv *BlockVolume) HaCount() int { return gv.ha }
