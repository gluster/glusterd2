package glustershd

import (
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"
)

type actionType uint16

const (
	actionStart actionType = iota
	actionStop
)

func selfhealdAction(c transaction.TxnCtx, action actionType) error {
	glustershDaemon, err := newGlustershd()
	if err != nil {
		return err
	}

	switch action {
	case actionStart:
		err = daemon.Start(glustershDaemon, true)
		if err == errors.ErrProcessAlreadyRunning {

			c.Logger().WithError(err).Error("Seld Heal Daemon is already running.")

			return nil
		}
	case actionStop:

		isVolRunning, err := volume.AreReplicateVolumesRunning()
		if err != nil {

			c.Logger().WithError(err).Error("Failed to get volinfo. Etcd server might be down.")

			return err
		}
		if !isVolRunning {
			err = daemon.Stop(glustershDaemon, true)
		}
	}

	return err
}

func txnSelfHealStart(c transaction.TxnCtx) error {
	return selfhealdAction(c, actionStart)

}

func txnSelfHealStop(c transaction.TxnCtx) error {
	return selfhealdAction(c, actionStop)
}
