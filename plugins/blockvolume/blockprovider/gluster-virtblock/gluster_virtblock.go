package glustervirtblock

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/utils"
	"github.com/gluster/glusterd2/plugins/blockvolume/blockprovider"
	"github.com/gluster/glusterd2/plugins/blockvolume/hostvol"

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
					log.WithError(err).Error("failed to create mount point")
					return "", err
				}
				notMnt = true
			} else {
				log.WithError(err).Error("failed to mount block host volume")
				return "", err
			}
		}

		if notMnt {
			err = volume.MountVolume(hostVolume, hostDir, "")
			if err != nil {
				log.WithError(err).Error("failed to mount block host volume")
				return "", err
			}
		}
		g.mounts[hostVolume] = hostDir
	}
	return hostDir, nil
}

// CreateBlockVolume will create a gluster block volume with given name and size having `hostVolume` as hosting volume
func (g *GlusterVirtBlk) CreateBlockVolume(name string, size uint64, hostVolume string, options ...blockprovider.BlockVolOption) (blockprovider.BlockVolume, error) {
	var (
		blockVolOpts = &blockprovider.BlockVolumeOptions{}
		clusterLocks = transaction.Locks{}
	)

	blockVolOpts.ApplyOpts(options...)
	logger := log.WithFields(log.Fields{
		"block_name":           name,
		"hostvol":              hostVolume,
		"requested_block_size": size,
	})

	// TODO: Check if block name already exists?
	hostDir, err := mountHost(g, hostVolume)
	if err != nil {
		logger.WithError(err).Error("failed to mount block hosting volume")
		return nil, err
	}

	blockFileName := hostDir + "/" + name
	err = utils.ExecuteCommandRun("truncate", fmt.Sprintf("-s %d", size), blockFileName) //nolint: gosec
	if err != nil {
		logger.WithError(err).Errorf("failed to truncate block file %s", blockFileName)
		return nil, err
	}

	if blockVolOpts.BlockType != "raw" {
		fsType := blockVolOpts.BlockType
		err = utils.ExecuteCommandRun(fmt.Sprintf("mkfs.%s", fsType), "-f", blockFileName) //nolint: gosec
		if err != nil {
			logger.WithError(err).Errorf("failed to format block file %s with filesystem %s", blockFileName, fsType)
			return nil, err
		}
	}

	resizeFunc := func(blockHostingAvailableSize, blockSize uint64) uint64 { return blockHostingAvailableSize - blockSize }
	if err = hostvol.ResizeBlockHostingVolume(hostVolume, size, resizeFunc); err != nil {
		logger.WithError(err).Error("failed in updating hostvolume _block-hosting-available-size metadata")
		return nil, err
	}

	if err = clusterLocks.Lock(hostVolume); err != nil {
		logger.WithError(err).Error("error in acquiring cluster lock")
		return nil, err
	}
	defer clusterLocks.UnLock(context.Background())

	volInfo, err := volume.GetVolume(hostVolume)
	if err != nil {
		logger.WithError(err).Errorf("failed to get host volume info %s", hostVolume)
		return nil, err
	}
	key := volume.BlockPrefix + name
	val := strconv.FormatUint(size, 10)
	volInfo.Metadata[key] = val
	if err := volume.AddOrUpdateVolume(volInfo); err != nil {
		logger.WithError(err).Error("failed in updating volume info to store")
		return nil, err
	}

	return &BlockVolume{
		hostVolume: hostVolume,
		name:       name,
		size:       size,
	}, nil
}

func delBlockEntry(hostName string, name string) error {
	var (
		clusterLocks = transaction.Locks{}
	)

	if err := clusterLocks.Lock(hostName); err != nil {
		log.WithError(err).Error("error in acquiring cluster lock")
		return err
	}
	defer clusterLocks.UnLock(context.Background())

	hostVol, err := volume.GetVolume(hostName)
	if err != nil {
		log.WithError(err).Error("failed to get block host vol info")
		return err
	}

	for k := range hostVol.Metadata {
		if k == (volume.BlockPrefix + name) {
			delete(hostVol.Metadata, k)
			if err := volume.AddOrUpdateVolume(hostVol); err != nil {
				log.WithError(err).Error("failed in updating volume info to store")
				return err
			}
		}
	}
	return nil
}

// DeleteBlockVolume deletes a gluster block volume of give name
func (g *GlusterVirtBlk) DeleteBlockVolume(name string, options ...blockprovider.BlockVolOption) error {
	var (
		blockVolOpts = &blockprovider.BlockVolumeOptions{}
	)

	blockVolOpts.ApplyOpts(options...)

	blkVol, err := g.GetBlockVolume(name)
	if err != nil || blkVol == nil {
		return errors.ErrBlockVolNotFound
	}

	hostDir, err := mountHost(g, blkVol.HostVolume())
	if err != nil {
		log.WithError(err).Errorf("error mounting block hosting volume :%s", blkVol.HostVolume())
		return err
	}

	blockFileName := hostDir + "/" + name
	err = os.Remove(blockFileName)
	if err != nil {
		log.WithError(err).Errorf("error removing block :%s", blockFileName)
		return err
	}

	err = delBlockEntry(blkVol.HostVolume(), name)
	if err != nil {
		log.WithError(err).Error("error updating block host volume metadata")
		return err
	}

	resizeFunc := func(blockHostingAvailableSize, blockSize uint64) uint64 { return blockHostingAvailableSize + blockSize }
	if err = hostvol.ResizeBlockHostingVolume(blkVol.HostVolume(), blkVol.Size(), resizeFunc); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"size":  blkVol.Size(),
		}).Error("error in resizing the block hosting volume")
	}

	return err
}

// GetBlockVolume gives info about a gluster block volume
func (g *GlusterVirtBlk) GetBlockVolume(name string) (blockprovider.BlockVolume, error) {
	volumes, err := volume.GetVolumes(context.Background())
	if err != nil {
		return nil, errors.ErrBlockHostVolNotFound
	}
	volumes = volume.ApplyFilters(volumes, volume.BlockHosted)

	for _, vols := range volumes {
		for k, v := range vols.Metadata {
			if k == (volume.BlockPrefix + name) {
				size, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return nil, err
				}
				return &BlockVolume{
					name:       name,
					size:       size,
					hostVolume: vols.Name,
				}, nil
			}
		}
	}

	return nil, errors.ErrBlockVolNotFound
}

// BlockVolumes returns all available gluster block volume
func (g *GlusterVirtBlk) BlockVolumes() []blockprovider.BlockVolume {
	var glusterBlockVolumes = []blockprovider.BlockVolume{}

	volumes, err := volume.GetVolumes(context.Background())
	if err != nil {
		return glusterBlockVolumes
	}

	volumes = volume.ApplyFilters(volumes, volume.BlockHosted)

	for _, vols := range volumes {
		for k, v := range vols.Metadata {
			if strings.Contains(k, volume.BlockPrefix) {
				blkSize, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return glusterBlockVolumes
				}
				blkName := strings.TrimPrefix(k, volume.BlockPrefix)
				glusterBlockVolumes = append(glusterBlockVolumes, &BlockVolume{name: blkName, hostVolume: vols.Name, size: blkSize})
			}
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
