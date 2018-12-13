package glustershd

import (
	"os"
	"path"

	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	log "github.com/sirupsen/logrus"
	config "github.com/spf13/viper"
)

type shdActor struct{}

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
		if val, ok := v.Options[shdKey]; ok && val == "off" {
			return nil
		}

		err = volgen.ClusterVolfileToFile(v, glustershDaemon.VolfileID, "glustershd")
		if err != nil {
			return err
		}

		err = daemon.Start(glustershDaemon, true, logger)
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
		if key != selfHealKey && key != granularEntryHealKey {
			return nil
		}
		switch key {
		case selfHealKey:
			glustershDaemon, err := newGlustershd()
			if err != nil {
				return err
			}

			switch value {
			case "on":
				if isHealEnabled(v) {
					err = volgen.ClusterVolfileToFile(v, glustershDaemon.VolfileID, "glustershd")
					if err != nil {
						return err
					}
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
		case granularEntryHealKey:
			switch value {
			case "enable":
				glusterdSockpath := path.Join(config.GetString("rundir"), "glusterd2.socket")
				options := []string{"granular-entry-heal-op", "glusterd-sock", glusterdSockpath}
				_, err := runGlfshealBin(v.Name, options)
				if err != nil {
					return err
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
		if val, ok := v.Options[shdKey]; ok && val == "off" {
			return nil
		}

		isVolRunning, err := volume.AreReplicateVolumesRunning(v.ID)
		if err != nil {
			return err
		}
		if !isVolRunning {
			return daemon.Stop(glustershDaemon, true, logger)
		}

	case xlator.VolumeStop:
		err = volgen.ClusterVolfileToFile(v, glustershDaemon.VolfileID, "glustershd")
		if err != nil {
			return err
		}
		err = daemon.Start(glustershDaemon, true, logger)
		if err != nil && err != gderrors.ErrProcessAlreadyRunning {
			return err
		}

	case xlator.VolumeSet:
		fallthrough
	case xlator.VolumeReset:
		if key != selfHealKey && key != granularEntryHealKey {
			return nil
		}
		switch key {
		case selfHealKey:
			glustershDaemon, err := newGlustershd()
			if err != nil {
				return err
			}

			switch value {
			case "off":
				if isHealEnabled(v) {
					err = volgen.ClusterVolfileToFile(v, glustershDaemon.VolfileID, "glustershd")
					if err != nil {
						return err
					}
					err = daemon.Start(glustershDaemon, true, logger)
					if err != nil && err != gderrors.ErrProcessAlreadyRunning {
						return err
					}
				}
			case "on":
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
		case granularEntryHealKey:
			switch value {
			case "disable":
				glusterdSockpath := path.Join(config.GetString("rundir"), "glusterd2.socket")
				options := []string{"granular-entry-heal-op", "glusterd-sock", glusterdSockpath}
				_, err := runGlfshealBin(v.Name, options)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil
}

func init() {
	xlator.RegisterOptionActor("replicate", &shdActor{})
}
