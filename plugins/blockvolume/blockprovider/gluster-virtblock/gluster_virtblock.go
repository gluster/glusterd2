package glustervirtblock

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"
	"github.com/gluster/glusterd2/plugins/blockvolume/blockprovider"
	blkUtils "github.com/gluster/glusterd2/plugins/blockvolume/utils"

	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
	"k8s.io/kubernetes/pkg/util/mount"
)

const providerName = "virtblock"

var mounter = mount.New("")

func init() {
	blockprovider.RegisterBlockProvider(providerName, newGlusterVirtBlk)
}

// GlusterVirtBlk implements block Provider interface. It represents a gluster-block
type GlusterVirtBlk struct {
	mounts map[string]string
}

func newGlusterVirtBlk() (blockprovider.Provider, error) {
	gb := &GlusterVirtBlk{}

	gb.mounts = make(map[string]string)

	return gb, nil
}

func mountHost(g *GlusterVirtBlk, hostVolume string) (string, error) {
	hostDir := g.mounts[hostVolume]
	if hostDir == "" {
		hostDir = config.GetString("rundir") + "/blockvolume/" + hostVolume
		notMnt, err := mounter.IsLikelyNotMountPoint(hostDir)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(hostDir, os.ModeDir|os.ModePerm)
				if err != nil {
					return "", fmt.Errorf("failed to create mount point %+v", err)
				}
				notMnt = true
			} else {
				return "", fmt.Errorf("failed to mount block host volume %+v", err)
			}
		}

		if notMnt {
			err = volume.MountVolume(hostVolume, hostDir, "")
			if err != nil {
				return "", fmt.Errorf("failed to mount block host volume %+v", err)
			}
		}
		g.mounts[hostVolume] = hostDir
	}
	return hostDir, nil
}

// CreateBlockVolume will create a gluster block volume with given name and size having `hostVolume` as hosting volume
func (g *GlusterVirtBlk) CreateBlockVolume(name string, size uint64, hostVolume string, options ...blockprovider.BlockVolOption) (blockprovider.BlockVolume, error) {
	blockVolOpts := &blockprovider.BlockVolumeOptions{}
	blockVolOpts.ApplyOpts(options...)
	logger := log.WithFields(log.Fields{
		"block_name":           name,
		"hostvol":              hostVolume,
		"requested_block_size": size,
	})

	hostDir, err := mountHost(g, hostVolume)
	if err != nil {
		return nil, fmt.Errorf("failed to mount block hosting volume %+v", err)
	}

	blockFileName := hostDir + "/" + name
	err = utils.ExecuteCommandRun("truncate", fmt.Sprintf("-s %d", size), blockFileName) //nolint: gosec
	if err != nil {
		return nil, fmt.Errorf("failed to truncate block file %s: %+v", blockFileName, err)
	}

	err = utils.ExecuteCommandRun("mkfs.xfs", "-f", blockFileName) //nolint: gosec
	if err != nil {
		return nil, fmt.Errorf("failed to format block file %s: %+v", blockFileName, err)
	}

	resizeFunc := func(blockHostingAvailableSize, blockSize uint64) uint64 { return blockHostingAvailableSize - blockSize }
	if err = blkUtils.ResizeBlockHostingVolume(hostVolume, size, resizeFunc); err != nil {
		logger.WithError(err).Error("failed in updating hostvolume _block-hosting-available-size metadata")
	}

	return &BlockVolume{
		hostVolume: hostVolume,
		name:       name,
		size:       size,
	}, err
}

// DeleteBlockVolume deletes a gluster block volume of give name
func (g *GlusterVirtBlk) DeleteBlockVolume(name string, options ...blockprovider.BlockVolOption) error {
	var (
		blockVolOpts = &blockprovider.BlockVolumeOptions{}
		hostVol      string
	)

	blockVolOpts.ApplyOpts(options...)

	// TODO: Listing all the block volumes to delete one block vol will bottleneck at scale. Possible options:
	// - Let block delete carry the host volume(optionally). The caller needs to keep this info returned in create vol, and send it in delete req.
	// - Build a map in memory ([blockvolume]hostvolume)during init(or lazy) during init of provider/create of block volume
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

	hostDir, err := mountHost(g, hostVol)
	if err != nil {
		return err
	}

	blockFileName := hostDir + "/" + name
	stat, err := os.Stat(blockFileName)
	if err != nil {
		return err
	}

	err = os.Remove(blockFileName)
	if err != nil {
		return err
	}

	size := stat.Size()
	resizeFunc := func(blockHostingAvailableSize, blockSize uint64) uint64 { return blockHostingAvailableSize + blockSize }
	if err = blkUtils.ResizeBlockHostingVolume(hostVol, size, resizeFunc); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"size":  size,
		}).Error("error in resizing the block hosting volume")
	}

	return err
}

// GetBlockVolume gives info about a gluster block volume
func (g *GlusterVirtBlk) GetBlockVolume(name string) (blockprovider.BlockVolume, error) {
	var (
		blockVolume           blockprovider.BlockVolume
		availableBlockVolumes = g.BlockVolumes()
	)

	//TODO: looping through all block volumes to get one block vol info is not scalable, fix it
	for _, blockVol := range availableBlockVolumes {
		if blockVol.Name() == name {
			blockVolume = blockVol
			break
		}
	}

	if blockVolume == nil {
		return nil, errors.New("block volume not found")
	}

	glusterBlockVol := &BlockVolume{
		name:       blockVolume.Name(),
		hostVolume: blockVolume.HostVolume(),
		size:       blockVolume.Size(),
	}

	return glusterBlockVol, nil
}

// BlockVolumes returns all available gluster block volume
func (g *GlusterVirtBlk) BlockVolumes() []blockprovider.BlockVolume {
	var glusterBlockVolumes = []blockprovider.BlockVolume{}

	volumes, err := volume.GetVolumes(context.Background())
	if err != nil {
		return glusterBlockVolumes
	}

	volumes = volume.ApplyFilters(volumes, volume.BlockHosted)

	for _, hostVol := range volumes {
		hostDir, err := mountHost(g, hostVol.Name)
		if err != nil {
			return glusterBlockVolumes
		}

		dirent, err := ioutil.ReadDir(hostDir)
		if err != nil {
			return glusterBlockVolumes
		}

		for _, blockVol := range dirent {
			glusterBlockVolumes = append(glusterBlockVolumes, &BlockVolume{name: blockVol.Name(), hostVolume: hostVol.Name, size: uint64(blockVol.Size())})
		}
	}

	return glusterBlockVolumes
}

// ProviderName returns name of block provider
func (g *GlusterVirtBlk) ProviderName() string {
	return providerName
}

// BlockVolume implements blockprovider.BlockVolume interface.
// It holds information about a gluster-block volume
type BlockVolume struct {
	hosts      []string
	hostVolume string
	name       string
	size       uint64
}

// HostAddresses returns host addresses of a gluster block vol
func (gv *BlockVolume) HostAddresses() []string { return gv.hosts }

// IQN returns IQN of a gluster block vol
func (gv *BlockVolume) IQN() string { return "" }

// Username returns username of a gluster-block vol.
func (gv *BlockVolume) Username() string { return "" }

// Password returns password for a gluster block vol
func (gv *BlockVolume) Password() string { return "" }

// HostVolume returns host vol name of gluster block
func (gv *BlockVolume) HostVolume() string { return gv.hostVolume }

// Name returns name of gluster block vol
func (gv *BlockVolume) Name() string { return gv.name }

// Size returns size of a gluster block vol in bytes
func (gv *BlockVolume) Size() uint64 { return gv.size }

// ID returns Gluster Block ID
func (gv *BlockVolume) ID() string { return "" }

// HaCount returns high availability count
func (gv *BlockVolume) HaCount() int { return 0 }
