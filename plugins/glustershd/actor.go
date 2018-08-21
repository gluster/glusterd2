package glustershd

import (
	"errors"
	"fmt"
	"os"

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

func getSelfHealKeys() []string {
	var selfhealKeys = make([]string, len(names))
	for i, n := range names {
		selfhealKeys[i] = fmt.Sprintf("%s.%s", n, selfHealKey)
	}
	return selfhealKeys
}

func (actor *shdActor) Do(v *volume.Volinfo, key string, value string, volOp xlator.VolumeOpType, logger log.FieldLogger) error {

	if v.Type != volume.Replicate && v.Type != volume.Disperse {
		return nil
	}

	glustershDaemon, err := newGlustershd()
	if err != nil {
		return err
	}
	switch volOp {
	case xlator.VolumeStart:
		for _, key := range getSelfHealKeys() {
			if val, ok := v.Options[key]; ok && val == "off" {
				return nil
			}
		}
		err := daemon.Start(glustershDaemon, true, logger)
		if err != nil && err != gderrors.ErrProcessAlreadyRunning {
			return err
		}

	case xlator.VolumeStop:
		isVolRunning, err := volume.AreReplicateVolumesRunning(v.ID)
		if err != nil {
			return err
		}

		if !isVolRunning {
			err := daemon.Stop(glustershDaemon, true, logger)
			if err != nil && err != gderrors.ErrPidFileNotFound {
				return err
			}
		}

	case xlator.VolumeSet:
		fallthrough
	case xlator.VolumeReset:
		if key != selfHealKey {
			return nil
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
				isVolRunning, err := volume.AreReplicateVolumesRunning(v.ID)
				if err != nil {
					return err
				}
				if !isVolRunning {
					err := daemon.Stop(glustershDaemon, true, logger)
					if !os.IsNotExist(err) {
						return err
					}
				}
			}

		}
	}
	return nil
}

func (actor *shdActor) Undo(v *volume.Volinfo, key string, value string, volOp xlator.VolumeOpType, logger log.FieldLogger) error {

	if v.Type != volume.Replicate && v.Type != volume.Disperse {
		return nil
	}

	glustershDaemon, err := newGlustershd()
	if err != nil {
		return err
	}
	switch volOp {
	case xlator.VolumeStart:
		for _, key := range getSelfHealKeys() {
			if val, ok := v.Options[key]; ok && val == "off" {
				return nil
			}
		}

		isVolRunning, err := volume.AreReplicateVolumesRunning(v.ID)
		if err != nil {
			return err
		}
		if !isVolRunning {
			return daemon.Stop(glustershDaemon, true, logger)
		}

	case xlator.VolumeStop:
		err = daemon.Start(glustershDaemon, true, logger)
		if err != nil && err != gderrors.ErrProcessAlreadyRunning {
			return err
		}

	case xlator.VolumeSet:
		fallthrough
	case xlator.VolumeReset:
		if key != selfHealKey {
			return nil
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
				isVolRunning, err := volume.AreReplicateVolumesRunning(v.ID)
				if err != nil {
					return err
				}
				if !isVolRunning {
					return daemon.Stop(glustershDaemon, true, logger)
				}
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
