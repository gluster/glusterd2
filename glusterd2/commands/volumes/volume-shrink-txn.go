package volumecommands

import (
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	rebalance "github.com/gluster/glusterd2/plugins/rebalance"
	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"
)

func startRebalance(c transaction.TxnCtx) error {
	var rinfo rebalanceapi.RebalInfo
	err := c.Get("rinfo", &rinfo)
	if err != nil {
		return err
	}

	rebalanceProcess, err := rebalance.NewRebalanceProcess(rinfo)
	if err != nil {
		return err
	}

	err = daemon.Start(rebalanceProcess, true, c.Logger())
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", rinfo.Volname).Error("Starting rebalance process failed")
		return err
	}

	return err
}
