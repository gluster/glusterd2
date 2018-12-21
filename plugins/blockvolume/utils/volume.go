package utils

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
	"github.com/gluster/glusterd2/glusterd2/transaction"
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
func GetExistingBlockHostingVolume(size uint64) (*volume.Volinfo, error) {
	var (
		filters     = []volume.Filter{volume.FilterBlockHostedVolumes, BlockSizeFilter(size)}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	)

	defer cancel()

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

// ResizeBlockHostingVolume will adds deletedBlockSize to block-hosting-available-size
// in metadata and update the new vol info to store.
func ResizeBlockHostingVolume(volname string, deletedBlockSize string) error {
	clusterLocks := transaction.Locks{}

	if err := clusterLocks.Lock(volname); err != nil {
		log.WithError(err).Error("error in acquiring cluster lock")
		return err
	}
	defer clusterLocks.UnLock(context.Background())

	volInfo, err := volume.GetVolume(volname)
	if err != nil {
		return err
	}

	deletedSize, err := size.Parse(deletedBlockSize)
	if err != nil {
		return err
	}

	if _, found := volInfo.Metadata[volume.BlockHostingAvailableSize]; !found {
		return errors.New("block-hosting-available-size metadata not found for volume")
	}

	availableSizeInBytes, err := strconv.ParseUint(volInfo.Metadata[volume.BlockHostingAvailableSize], 10, 64)
	if err != nil {
		return err
	}

	volInfo.Metadata[volume.BlockHostingAvailableSize] = fmt.Sprintf("%d", availableSizeInBytes+uint64(deletedSize))

	return volume.AddOrUpdateVolume(volInfo)
}

// SelectRandomVolume will select a random volume from a given slice of volumes
func SelectRandomVolume(volumes []*volume.Volinfo) (*volume.Volinfo, error) {
	if len(volumes) == 0 {
		return nil, errors.New("no available volumes")
	}

	i := rand.Int() % len(volumes)
	return volumes[i], nil
}
