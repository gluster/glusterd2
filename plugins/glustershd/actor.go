package glustershd

import (
	"errors"

	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	log "github.com/sirupsen/logrus"
)

var names = [...]string{"replicate", "afr"}

const (
	selfHealKey = "self-heal-daemon"
)

type shdActor struct{}

func (actor *shdActor) Do(v *volume.Volinfo, key string, value string, logger log.FieldLogger) error {
	if key != selfHealKey {
		return nil
	}
	glustershDaemon, err := newGlustershd()
	if err != nil {
		return err
	}

	switch value {
	case "on":
		if isHealEnabled(v) {
			err = daemon.Start(glustershDaemon, true, logger)
			if err != nil && err != gderrors.ErrProcessAlreadyRunning {
				return err
			}
		}
	case "off":
		if !isHealEnabled(v) {
			isVolRunning, err := volume.AreReplicateVolumesRunning()
			if err != nil {
				return err
			}
			if !isVolRunning {
				return daemon.Stop(glustershDaemon, true, logger)
			}
		}
	}
	return nil
}

func (actor *shdActor) Undo(v *volume.Volinfo, key string, value string, logger log.FieldLogger) error {
	if key != selfHealKey {
		return nil
	}

	glustershDaemon, err := newGlustershd()
	if err != nil {
		return err
	}

	switch value {
	case "off":
		if !isHealEnabled(v) {
			if v.State != volume.VolStarted {
				return errors.New("volume should be in started state")
			}
			err = daemon.Start(glustershDaemon, true, logger)
			if err != nil && err != gderrors.ErrProcessAlreadyRunning {
				return err
			}
		}
	case "on":
		if isHealEnabled(v) {
			isVolRunning, err := volume.AreReplicateVolumesRunning()
			if err != nil {
				return err
			}
			if !isVolRunning {
				return daemon.Stop(glustershDaemon, true, logger)
			}
		}
	}
	return nil
}

func init() {
	for _, name := range names {
		xlator.RegisterOptionActor(name, &shdActor{})
	}
}
