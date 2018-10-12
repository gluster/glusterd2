package utils

import (
	"context"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/provisioners"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/utils"

	log "github.com/sirupsen/logrus"
)

// MountLocalBricks mounts bricks of auto provisioned volumes
func MountLocalBricks() error {
	volumes, err := volume.GetVolumes(context.TODO())
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
			// Mount all local bricks if they are auto provisioned or inherited via snapshot creation
			provisionType := b.PType
			if provisionType.IsAutoProvisioned() || provisionType.IsSnapshotProvisioned() {
				mountRoot := strings.TrimSuffix(b.Path, b.MountInfo.Mountdir)
				if _, exists := mounts[mountRoot]; exists {
					continue
				}

				if v.Provisioner != "" {
					provisioner, err := provisioners.Get(v.Provisioner)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"volume":      v.Name,
							"provisioner": v.Provisioner,
						}).Error("unable to get provisioner")
						continue
					}

					err = provisioner.MountBrick(b.Device, b.Name, b.Path)
					if err != nil {
						log.WithError(err).WithFields(log.Fields{
							"volume": v.Name,
							"device": b.Device,
							"name":   b.Name,
							"path":   b.Path,
						}).Error("brick mount failed")
					}
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
