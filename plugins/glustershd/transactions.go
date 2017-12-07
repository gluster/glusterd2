package glustershd

import (
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"
)

func txnSelfHealStart(c transaction.TxnCtx) error {
	glustershDaemon, err := newGlustershd()
	if err != nil {
		return err
	}
	err = daemon.Start(glustershDaemon, true)
	return err
}
