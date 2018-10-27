package brickmux

import (
	"context"
	"os"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"

	log "github.com/sirupsen/logrus"
)

// Reconcile will multiplex bricks on (re)start of glusterd2.
func Reconcile() error {

	bmuxEnabled, err := Enabled()
	if err != nil {
		return err
	}

	if !bmuxEnabled {
		return nil
	}

	volumes, err := volume.GetVolumes(context.TODO())
	if err != nil {
		return err
	}

	var bricks []brick.Brickinfo
	for _, v := range volumes {
		if v.State == volume.VolStarted {
			bricks = append(bricks, v.GetLocalBricks()...)
		}
	}

	// convert it to map for easier lookup below
	volumeMap := make(map[string]*volume.Volinfo)
	for _, volume := range volumes {
		volumeMap[volume.ID.String()] = volume
	}

	for _, b := range bricks {
		d, err := brick.NewGlusterfsd(b)
		if err != nil {
			return err
		}

		if running, _ := daemon.IsRunning(d); running {
			continue
		} else {
			// cleanup stale pidfile
			os.Remove(d.PidFile())
		}

		err = Multiplex(b, volumeMap[b.VolumeID.String()], volumes, log.StandardLogger())
		switch err {
		case nil:
			// successfully multiplexed
			continue
		case ErrNoCompat:
			// something changed between restart and we can't find
			// a compatible brick process.
			// do nothing, fallback to starting a separate process
			log.WithField("brick", b.String()).Warn(err)
		default:
			// log and move on, do not exit; this behaviour can be changed later if necessarry
			log.WithField("brick", b.String()).WithError(err).Error("brickmux.Multiplex failed")
			continue
		}

		if err := b.StartBrick(log.StandardLogger()); err != nil {
			if err == errors.ErrProcessAlreadyRunning {
				continue
			}
			log.WithField("brick", b.String()).WithError(err).Error("failed to start brick process")
		}
	}

	return nil
}
