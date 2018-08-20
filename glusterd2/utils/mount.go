package utils

import (
	"strings"

	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"

	log "github.com/sirupsen/logrus"
)

// MountLocalBricks mounts bricks of auto provisioned volumes
func MountLocalBricks() error {
	volumes, err := volume.GetVolumes()
	if err != nil {
		return err
	}

	// TODO: Get Snapshot Volumes as well

	if len(volumes) == 0 {
		return nil
	}

	// Get list of mounted dirs
	mtabEntries, err := volume.GetMounts()
	if err != nil {
		log.WithError(err).Error("failed to get list of mounts")
		return err
	}

	mounts := make(map[string]struct{})

	for _, entry := range mtabEntries {
		mounts[entry.MntDir] = struct{}{}
	}

	for _, v := range volumes {
		for _, b := range v.GetLocalBricks() {
			// Mount all local Bricks if they are auto provisioned
			if b.MountInfo.DevicePath != "" {
				mountRoot := strings.TrimSuffix(b.Path, b.MountInfo.Mountdir)
				if _, exists := mounts[mountRoot]; exists {
					continue
				}

				err := utils.ExecuteCommandRun("mount", "-o", b.MountInfo.MntOpts, b.MountInfo.DevicePath, mountRoot)
				if err != nil {
					log.WithError(err).WithFields(log.Fields{
						"volume": v.Name,
						"dev":    b.MountInfo.DevicePath,
						"path":   mountRoot,
					}).Error("brick mount failed")
				}
			}
		}
	}

	return nil
}
