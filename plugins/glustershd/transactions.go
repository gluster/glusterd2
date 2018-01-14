package glustershd

import (
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
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
