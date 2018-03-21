package quota

import (
	"os"
	"path"

	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/glusterd2/xlator"
	"github.com/gluster/glusterd2/pkg/errors"

	log "github.com/sirupsen/logrus"
)

type quotadActor struct{}

const (
	quotaDaemonKey = "enable"
)

// isQuotadStopRequired checks if the quota daemon has to be stopped.
// The quotad process needs to be stopped on a peer only when that peer
// does not have bricks belonging to a quota enabled volume.
func isQuotadStopRequired(volumes []*volume.Volinfo) bool {
	for _, v := range volumes {
		if !isQuotaEnabled(v) {
			continue
		} else if v.State != volume.VolStarted {
			continue
		} else {
			bricks := v.GetLocalBricks()
			if len(bricks) > 0 {
				return false
			}
		}
	}
	return true
}

func (actor *quotadActor) Do(v *volume.Volinfo, key, value string, logger log.FieldLogger) error {
	if key != quotaDaemonKey {
		return nil
	}
	quotadDaemon, err := NewQuotad()
	if err != nil {
		return err
	}
	// Create pidfile dir if not exists
	if err := os.MkdirAll(path.Dir(quotadDaemon.pidfilepath),
		os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	// Create logFiledir dir
	if err := os.MkdirAll(path.Dir(quotadDaemon.logfilepath),
		os.ModeDir|os.ModePerm); err != nil {
		return err
	}

	volumes, err := volume.GetVolumes()
	if err != nil {
		logger.WithError(err).Error("failed to get volumes")
		return err
	}

	if value == "off" && isQuotadStopRequired(volumes) {
		// This condition is for disabling quotad
		if err = daemon.Stop(quotadDaemon, true, logger); err != nil {
			logger.Error("quotad stop failed")
		}
	} else {
		// Quotad must be restarted whenever quota is enabled or disabled
		// for a volume.
		err = daemon.Stop(quotadDaemon, true, logger)
		if err == errors.ErrPidFileNotFound {
			logger.Info("quotad stop failed as pidfile missing")
		} else if err != nil {
			logger.Warn("quotad stop failed")
		} else {
			logger.Info("quotad stopped for restart")
		}
		if err = daemon.Start(quotadDaemon, true, logger); err != nil {
			logger.WithError(err).Error("quotad start failed")
		}
	}
	return err
}

func (actor *quotadActor) Undo(v *volume.Volinfo, key, value string, logger log.FieldLogger) error {
	//nothing needs to be done as of now.
	return nil
}

func init() {
	xlator.RegisterOptionActor("quota", &quotadActor{})
}
