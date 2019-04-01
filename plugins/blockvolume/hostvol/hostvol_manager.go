package hostvol

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gluster/glusterd2/glusterd2/commands/volumes"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/plugins/blockvolume/api"
	config "github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

const (
	globalLockID = "host-vol-lock"
)

// HostingVolumeManager provides methods for host volume management
type HostingVolumeManager interface {
	GetHostingVolumesInUse() []*volume.Volinfo
	GetOrCreateHostingVolume(name string, blkName string, minSizeLimit uint64, hostVolumeInfo *api.HostVolumeInfo) (*volume.Volinfo, error)
	DeleteBlockInfoFromBHV(hostVol string, blkName string, size uint64) error
}

// GlusterVolManager is a concrete implementation of HostingVolumeManager
type GlusterVolManager struct {
	hostVolOpts *HostingVolumeOptions
}

// NewGlusterVolManager returns a glusterVolManager instance
func NewGlusterVolManager() *GlusterVolManager {
	g := &GlusterVolManager{
		hostVolOpts: newHostingVolumeOptions(),
	}

	return g
}

// GetHostingVolumesInUse lists all volumes which used in hosting block-vols
func (g *GlusterVolManager) GetHostingVolumesInUse() []*volume.Volinfo {
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
func (g *GlusterVolManager) GetOrCreateHostingVolume(name string, blkName string, minSizeLimit uint64, hostVolumeInfo *api.HostVolumeInfo) (*volume.Volinfo, error) {
	var (
		volInfo      *volume.Volinfo
		clusterLocks = transaction.Locks{}
	)

	if err := clusterLocks.Lock(path.Join(globalLockID, name)); err != nil {
		return nil, err
	}
	defer clusterLocks.UnLock(context.Background())

	g.hostVolOpts.SetFromClusterOptions()
	g.hostVolOpts.SetFromReq(hostVolumeInfo)
	volCreateReq, err := g.hostVolOpts.PrepareVolumeCreateReq()
	if err != nil {
		log.WithError(err).Error("failed to create block volume create request")
		return nil, err
	}

	// ERROR if HostingVolume is not specified and auto-create-block-hosting-volumes is false
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
	// TODO: Since this is not done within volume lock, this volumes' available size might have been
	// changed by the time we actually reserve the size in updateBhvInfoAndSize(). This can lead
	// updateBhvInfoAndSize() to fail with no space. We do not retry block create in this case,
	// the application can retry to workaround this race.
	if name == "" {
		vInfo, err := GetExistingBlockHostingVolume(minSizeLimit, g.hostVolOpts)
		if err != nil {
			log.WithError(err).Debug("no block hosting volumes present")
		}
		volInfo = vInfo
	}

	// If No volumes are available with Metadata:block-hosting=yes or if no space available to create block
	// volumes(Metadata:block-hosting-available-size is less than request size), then try to create a new
	// block hosting Volume with generated name with default size and volume type configured.
	if name == "" && volInfo == nil {
		vInfo, err := CreateAndStartHostingVolume(volCreateReq)
		if err != nil {
			log.WithError(err).Error("error in auto creation of block hosting volume")
			return nil, err
		}
		volInfo = vInfo
	}

	if err = clusterLocks.Lock(volInfo.Name); err != nil {
		log.WithError(err).Error("error in acquiring cluster lock")
		return nil, err
	}
	defer clusterLocks.UnLock(context.Background())

	volInfo, err = g.updateBhvInfoAndSize(volInfo.Name, blkName, minSizeLimit)
	if err != nil {
		log.WithError(err).Error("error in obtaining block host volume")
		return nil, err
	}

	return volInfo, nil
}

// updateBhvInfoAndSize will set the block host vol info in metadata and also reserve the size required for creating the new block in the input hostvolume
func (g *GlusterVolManager) updateBhvInfoAndSize(hostVolume string, blkName string, minSizeLimit uint64) (*volume.Volinfo, error) {

	volInfo, err := volume.GetVolume(hostVolume)
	if err != nil {
		log.WithError(err).Errorf("failed to get host volume info %s", hostVolume)
		return nil, err
	}

	if volInfo.Metadata[volume.BlockHostMarkedForPrune] == "true" {
		return nil, errors.New("block host volume masked for prune, retry")
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

	key := volume.BlockPrefix + blkName
	val := strconv.FormatUint(minSizeLimit, 10)
	volInfo.Metadata[key] = val

	resizeFunc := func(blockHostingAvailableSize, blockSize uint64) uint64 { return blockHostingAvailableSize - blockSize }
	if err = UpdateBlockHostingVolumeSize(volInfo, minSizeLimit, resizeFunc); err != nil {
		log.WithError(err).Error("failed in updating hostvolume _block-hosting-available-size metadata")
		return nil, err
	}

	// Note that any further error exit conditions should undo the above hostsize change
	if err := volume.AddOrUpdateVolume(volInfo); err != nil {
		log.WithError(err).Error("failed in updating volume info to store")
	}

	return volInfo, nil
}

// DeleteBlockInfoFromBHV resets the available space on the bhv and also deletes the block entry in the metadata of the bhv
// In this function, if the bhv is empty i.e. there are no blocks, then the bhv delete is initiated
func (g *GlusterVolManager) DeleteBlockInfoFromBHV(hostVol string, blkName string, size uint64) error {
	var (
		clusterLocks = transaction.Locks{}
		prune        = false
	)

	if err := clusterLocks.Lock(hostVol); err != nil {
		log.WithError(err).Error("error in acquiring cluster lock")
		return err
	}
	volInfo, err := volume.GetVolume(hostVol)
	if err != nil {
		log.WithError(err).Errorf("failed to get host volume info %s", hostVol)
		clusterLocks.UnLock(context.Background())
		return err
	}

	for k := range volInfo.Metadata {
		if k == (volume.BlockPrefix + blkName) {
			delete(volInfo.Metadata, k)
		}
	}

	resizeFunc := func(blockHostingAvailableSize, blockSize uint64) uint64 { return blockHostingAvailableSize + blockSize }
	if err = UpdateBlockHostingVolumeSize(volInfo, size, resizeFunc); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"size":  size,
		}).Error("error in resizing the block hosting volume")
	}

	// TODO: Also make sure volInfo.Metadata[volume.BlockPrefix*] has no keys left
	availableSizeInBytes, err := strconv.ParseUint(volInfo.Metadata[volume.BlockHostingAvailableSize], 10, 64)
	if err != nil {
		clusterLocks.UnLock(context.Background())
		return err
	}
	if availableSizeInBytes == volInfo.Capacity {
		if g.hostVolOpts.AutoDelete {
			volInfo.Metadata[volume.BlockHostMarkedForPrune] = "true"
			prune = true
		}
	}

	if err := volume.AddOrUpdateVolume(volInfo); err != nil {
		log.WithError(err).Error("failed in updating volume info to store")
		clusterLocks.UnLock(context.Background())
		return err
	}
	clusterLocks.UnLock(context.Background())

	if prune == true {
		err = g.pruneBHV(volInfo.Name, blkName, size)
		log.WithError(err).Errorf("failed to prune block host volume %s after deleting block %s", volInfo.Name, blkName)
	}

	return nil
}

// RegisterBHVstepFunctions registers the functions for the transaction
func RegisterBHVstepFunctions() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"bhv.unmount", BhvUnmount},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

// BhvUnmount unmount the block host volume
func BhvUnmount(c transaction.TxnCtx) error {
	var hostVol string
	if err := c.Get("bhvName", &hostVol); err != nil {
		return err
	}

	mntPath := path.Join(config.GetString("rundir"), "/blockvolume/", hostVol)
	_ = syscall.Unmount(mntPath, syscall.MNT_FORCE)

	return nil
}

// pruneBHV deletes the block host volume that is marked for deletion
func (g *GlusterVolManager) pruneBHV(hostVol string, blkName string, size uint64) error {
	var (
		ctx = gdctx.WithReqLogger(context.Background(), log.StandardLogger())
	)

	if !g.hostVolOpts.AutoDelete {
		return nil
	}

	logger := gdctx.GetReqLogger(ctx)
	logger.Info("Unmounting and deleting block host volume:%s", hostVol)

	allNodes, err := peer.GetPeerIDs()
	if err != nil {
		log.WithError(err).Error("error in getting peerIDs")
		return err
	}

	txn := transaction.NewTxn(ctx)
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "bhv.unmount",
			Nodes:  allNodes,
		},
	}
	txn.Ctx.Set("bhvName", hostVol)

	// Some nodes may not be up, which is okay.
	txn.DontCheckAlive = true
	txn.DisableRollback = true

	_ = txn.Do()
	txn.Done()

	_, _, err = volumecommands.StopVolume(ctx, hostVol)
	if err != nil {
		log.WithError(err).Error("error in stopping auto created block hosting volume")
		return err
	}

	_, _, err = volumecommands.DeleteVolume(ctx, hostVol)
	if err != nil {
		log.WithError(err).Error("error in auto deleting block hosting volume")
		return err
	}

	return nil
}
