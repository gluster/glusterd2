package glustershd

import (
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"
	log "github.com/sirupsen/logrus"
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
			log.Println("Self Heal Daemon is already running.")
			return nil
		}
	case actionStop:
		if !volume.AreReplicateVolumesRunning() {
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
