package utils

import (
	"context"

	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"

	log "github.com/sirupsen/logrus"
)

// MountLocalBricks mounts bricks of auto provisioned volumes
func MountLocalBricks() error {
	volumes, err := volume.GetVolumes(context.TODO())
	if err != nil {
		return err
	}
	snapVolumes, err := snapshot.GetActivatedSnapshotVolumes()
	if err != nil {
		return err
	}

	if len(snapVolumes) != 0 {
		volumes = append(volumes, snapVolumes...)
	} else if len(volumes) == 0 {
		return nil
	}

	for _, v := range volumes {
		if err := volume.MountVolumeBricks(v, true); err != nil {
			log.WithError(err).Error(errors.ErrVolumeBricksMountFailed)
			continue
		}
	}

	return nil
}
