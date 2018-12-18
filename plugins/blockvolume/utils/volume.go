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
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/size"

	log "github.com/sirupsen/logrus"
)

// BlockSizeFilter returns a volume Filter, which will filter out volumes
// haing block-hosting-available-size greater than give size.
func BlockSizeFilter(size int64) volume.Filter {
	return func(volinfos []*volume.Volinfo) []*volume.Volinfo {
		var volumes []*volume.Volinfo

		for _, volinfo := range volinfos {
			availableSize, found := volinfo.Metadata["block-hosting-available-size"]
			if !found {
				continue
			}

			if availableSizeInBytes, err := strconv.Atoi(availableSize); err == nil && int64(availableSizeInBytes) > size {
				volumes = append(volumes, volinfo)
			}
		}
		return volumes
	}
}

// GetExistingBlockHostingVolume returns a existing volume which is suitable for hosting a gluster-block
func GetExistingBlockHostingVolume(size int64) (*volume.Volinfo, error) {
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

// CreateBlockHostingVolume will create a gluster volume with metadata block-hosting-volume-auto-created=yes
func CreateBlockHostingVolume(req *api.VolCreateReq) (*volume.Volinfo, error) {
	status, err := volumecommands.CreateVolume(context.Background(), *req)
	if err != nil || status != http.StatusCreated {
		return nil, err
	}

	vInfo, err := volume.GetVolume(req.Name)
	if err != nil {
		return nil, err
	}

	vInfo.Metadata["block-hosting-volume-auto-created"] = "yes"
	return vInfo, nil
}

// ResizeBlockHostingVolume will adds deletedBlockSize to block-hosting-available-size
// in metadata and update the new vol info to store.
func ResizeBlockHostingVolume(volname string, deletedBlockSize string) error {
	clusterLocks := transaction.Locks{}

	volInfo, err := volume.GetVolume(volname)
	if err != nil {
		return err
	}

	deletedSizeInBytes, err := size.Parse(deletedBlockSize)
	if err != nil {
		return err
	}

	if _, found := volInfo.Metadata["block-hosting-available-size"]; !found {
		return errors.New("block-hosting-available-size metadata not found for volume")
	}

	availableSizeInBytes, err := strconv.Atoi(volInfo.Metadata["block-hosting-available-size"])
	if err != nil {
		return err
	}

	volInfo.Metadata["block-hosting-available-size"] = fmt.Sprintf("%d", size.Size(availableSizeInBytes)+deletedSizeInBytes)

	if err := clusterLocks.Lock(volInfo.Name); err != nil {
		log.WithError(err).Error("error in acquiring cluster lock")
		return err
	}

	defer clusterLocks.UnLock(context.Background())

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
