package blockvolume

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/plugins/blockvolume/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	globalLockID = "host-vol-lock"
)

// HostingVolumeManager provides methods for host volume management
type HostingVolumeManager interface {
	GetHostingVolumesInUse() []*volume.Volinfo
	GetOrCreateHostingVolume(name string, minSizeLimit uint64) (*volume.Volinfo, error)
}

// glusterVolManager is a concrete implementation of HostingVolumeManager
type glusterVolManager struct {
	hostVolOpts *HostingVolumeOptions
}

// newGlusterVolManager returns a glusterVolManager instance
func newGlusterVolManager() *glusterVolManager {
	var (
		g           = &glusterVolManager{}
		hostVolOpts = &HostingVolumeOptions{}
	)

	hostVolOpts.ApplyFromConfig(viper.GetViper())
	g.hostVolOpts = hostVolOpts
	return g
}

// GetHostingVolumesInUse lists all volumes which used in hosting block-vols
func (g *glusterVolManager) GetHostingVolumesInUse() []*volume.Volinfo {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	volumes, err := volume.GetVolumes(ctx)
	if err != nil || len(volumes) == 0 {
		return nil
	}

	return volume.ApplyFilters(volumes, volume.BlockHosted)
}

// GetOrCreateHostingVolume will returns volume details for a given volume name and having a minimum size of `minSizeLimit`.
// If volume name is not provided then it will create a gluster volume with default size for hosting gluster block.
func (g *glusterVolManager) GetOrCreateHostingVolume(name string, minSizeLimit uint64) (*volume.Volinfo, error) {
	var (
		volInfo      *volume.Volinfo
		volCreateReq = g.hostVolOpts.PrepareVolumeCreateReq()
		clusterLocks = transaction.Locks{}
	)

	if err := clusterLocks.Lock(path.Join(globalLockID, name)); err != nil {
		return nil, err
	}
	defer clusterLocks.UnLock(context.Background())

	// ERROR if If HostingVolume is not specified and auto-create-block-hosting-volumes is false
	if name == "" && !g.hostVolOpts.AutoCreate {
		err := errors.New("host volume is not provided and auto creation is not enabled")
		log.WithError(err).Error("failed in creating block volume")
		return nil, err
	}

	// If HostingVolume name is not empty, then create block volume with requested size.
	// If available size is less than requested size then ERROR. Set block related
	// metadata and volume options if not exists.
	if name != "" {
		vInfo, err := volume.GetVolume(name)
		if err != nil {
			log.WithError(err).Error("error in fetching volume info")
			return nil, err
		}
		volInfo = vInfo
	}

	// If HostingVolume is not specified. List all available volumes and see if any volume is
	// available with Metadata:block-hosting=yes
	if name == "" {
		vInfo, err := utils.GetExistingBlockHostingVolume(minSizeLimit)
		if err != nil {
			log.WithError(err).Debug("no block hosting volumes present")
		}
		volInfo = vInfo
	}

	// If No volumes are available with Metadata:block-hosting=yes or if no space available to create block
	// volumes(Metadata:block-hosting-available-size is less than request size), then try to create a new
	// block hosting Volume with generated name with default size and volume type configured.
	if name == "" && volInfo == nil {
		vInfo, err := utils.CreateAndStartHostingVolume(volCreateReq)
		if err != nil {
			log.WithError(err).Error("error in auto creation of block hosting volume")
			return nil, err
		}
		volInfo = vInfo
	}

	if _, found := volInfo.Metadata[volume.BlockHosting]; !found {
		volInfo.Metadata[volume.BlockHosting] = "yes"
	}

	blockHosting := volInfo.Metadata[volume.BlockHosting]

	if strings.ToLower(blockHosting) != "yes" {
		return nil, errors.New("not a block hosting volume")
	}

	if _, found := volInfo.Metadata[volume.BlockHostingAvailableSize]; !found {
		volInfo.Metadata[volume.BlockHostingAvailableSize] = fmt.Sprintf("%d", g.hostVolOpts.Size)
	}

	availableSizeInBytes, err := strconv.ParseUint(volInfo.Metadata[volume.BlockHostingAvailableSize], 10, 64)

	if err != nil {
		return nil, err
	}

	if availableSizeInBytes < minSizeLimit {
		return nil, fmt.Errorf("available size is less than requested size,request size: %d, available size: %d", minSizeLimit, availableSizeInBytes)
	}

	if volInfo.State != volume.VolStarted {
		return nil, errors.New("volume has not been started")
	}

	if err := volume.AddOrUpdateVolume(volInfo); err != nil {
		log.WithError(err).Error("failed in updating volume info to store")
	}

	return volInfo, nil
}
