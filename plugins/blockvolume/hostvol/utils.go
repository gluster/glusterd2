package hostvol

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gluster/glusterd2/glusterd2/commands/volumes"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/oldtransaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/size"
	log "github.com/sirupsen/logrus"
)

// BlockSizeFilter returns a volume Filter, which will filter out volumes
// haing block-hosting-available-size greater than give size.
func BlockSizeFilter(size uint64) volume.Filter {
	return func(volinfos []*volume.Volinfo) []*volume.Volinfo {
		var volumes []*volume.Volinfo

		for _, volinfo := range volinfos {
			availableSize, found := volinfo.Metadata[volume.BlockHostingAvailableSize]
			if !found {
				continue
			}

			if availableSizeInBytes, err := strconv.ParseUint(availableSize, 10, 64); err == nil && availableSizeInBytes > size {
				volumes = append(volumes, volinfo)
			}
		}
		return volumes
	}
}

// GetExistingBlockHostingVolume returns a existing volume which is suitable for hosting a gluster-block
func GetExistingBlockHostingVolume(size uint64, h *HostingVolumeOptions) (*volume.Volinfo, error) {
	var (
		filters     = []volume.Filter{volume.FilterBlockHostedVolumes, BlockSizeFilter(size)}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	)

	defer cancel()

	if h.ShardSize != 0 {
		filters = append(filters, volume.FilterShardVolumes)
	}
	if h.ThinArbPath != "" {
		filters = append(filters, volume.FilterThinArbiterVolumes)
	}

	volumes, err := volume.GetVolumes(ctx)
	if err != nil || len(volumes) == 0 {
		return nil, fmt.Errorf("%v/no volumes found", err)
	}

	volumes = volume.ApplyCustomFilters(volumes, filters...)

	return SelectRandomVolume(volumes)
}

// CreateAndStartHostingVolume creates and starts a gluster volume and returns volume Info on success.
// Set Metadata:block-hosting-volume-auto-created=yes if Block hosting volume is created and started successfully.
func CreateAndStartHostingVolume(req *api.VolCreateReq) (*volume.Volinfo, error) {
	ctx := gdctx.WithReqLogger(context.Background(), log.StandardLogger())

	status, err := volumecommands.CreateVolume(ctx, *req)
	if err != nil || status != http.StatusCreated {
		log.WithError(err).Error("error in auto creating block hosting volume")
		return nil, err
	}

	vInfo, _, err := volumecommands.StartVolume(ctx, req.Name, api.VolumeStartReq{})
	if err != nil {
		log.WithError(err).Error("error in starting auto created block hosting volume")
		return nil, err
	}

	vInfo.Metadata[volume.BlockHostingVolumeAutoCreated] = "yes"
	log.WithField("name", vInfo.Name).Debug("host volume created and started successfully")
	return vInfo, nil
}

// ResizeBlockHostingVolume will resize the _block-hosting-available-size metadata and update the new vol info to store.
// resizeFunc is use to update the new new value to the _block-hosting-available-size metadata
// e.g for adding the value to the _block-hosting-available-size metadata  we can use `resizeFunc` as
// f := func(a, b uint64) uint64{return a +b }
func ResizeBlockHostingVolume(volName string, blockSize interface{}, resizeFunc func(blockHostingAvailableSize, blockSize uint64) uint64) error {
	var (
		clusterLocks = oldtransaction.Locks{}
	)

	if err := clusterLocks.Lock(volName); err != nil {
		log.WithError(err).Error("error in acquiring cluster lock")
		return err
	}
	defer clusterLocks.UnLock(context.Background())

	volInfo, err := volume.GetVolume(volName)
	if err != nil {
		return err
	}

	err = UpdateBlockHostingVolumeSize(volInfo, blockSize, resizeFunc)
	if err != nil {
		return err
	}

	return volume.AddOrUpdateVolume(volInfo)
}

// UpdateBlockHostingVolumeSize will update the _block-hosting-available-size metadata in the volinfo passed
// resizeFunc is use to update the new new value to the _block-hosting-available-size metadata
// e.g for adding the value to the _block-hosting-available-size metadata  we can use `resizeFunc` as
// f := func(a, b uint64) uint64{return a +b }
func UpdateBlockHostingVolumeSize(volInfo *volume.Volinfo, blockSize interface{}, resizeFunc func(blockHostingAvailableSize, blockSize uint64) uint64) error {
	var (
		blkSize size.Size
	)

	if _, found := volInfo.Metadata[volume.BlockHostingAvailableSize]; !found {
		return errors.New("block-hosting-available-size metadata not found for volume")
	}

	availableSizeInBytes, err := strconv.ParseUint(volInfo.Metadata[volume.BlockHostingAvailableSize], 10, 64)
	if err != nil {
		return err
	}

	switch sz := blockSize.(type) {
	case string:
		blkSize, err = size.Parse(sz)
		if err != nil {
			return err
		}
	case uint64:
		blkSize = size.Size(sz)
	case int64:
		blkSize = size.Size(sz)
	default:
		return fmt.Errorf("blocksize is not a supported type(%T)", blockSize)
	}

	// TODO: If there are no blocks in the block hosting volume, delete the bhv?

	log.WithFields(log.Fields{
		"blockHostingAvailableSize": size.Size(availableSizeInBytes),
		"blockSize":                 blkSize,
	}).Debug("resizing hosting volume")

	volInfo.Metadata[volume.BlockHostingAvailableSize] = fmt.Sprintf("%d", resizeFunc(availableSizeInBytes, uint64(blkSize)))

	return nil
}

// SelectRandomVolume will select a random volume from a given slice of volumes
func SelectRandomVolume(volumes []*volume.Volinfo) (*volume.Volinfo, error) {
	if len(volumes) == 0 {
		return nil, errors.New("no available volumes")
	}

	i := rand.Int() % len(volumes)
	return volumes[i], nil
}
