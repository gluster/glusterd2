package rebalance

import (
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"
)

type actionType uint16

const (
	actionStart actionType = iota
	actionStop
)

func rebalanceCommand(c transaction.TxnCtx, cmd actionType) error {
	var rinfo rebalanceapi.RebalInfo
	err := c.Get("rinfo", &rinfo)
	if err != nil {
		return err
	}

	rebalanceProcess, err := NewRebalanceProcess(rinfo)
	if err != nil {
		return err
	}

	switch cmd {
	case actionStart:
		err = daemon.Start(rebalanceProcess, true)
		if err != nil {
			c.Logger().WithError(err).WithField(
				"volume", rinfo.Volname).Error("Starting rebalance process failed")
			return err
		}

	case actionStop:
		err = daemon.Stop(rebalanceProcess, true)
	}
	return err
}

func startRebalance(c transaction.TxnCtx) error {
	return rebalanceCommand(c, actionStart)
}

func stopRebalance(c transaction.TxnCtx) error {
	return rebalanceCommand(c, actionStop)
}
