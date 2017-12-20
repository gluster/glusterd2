package rebalance

import (
	//"fmt"

	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"
)

// RebalanceCommand represents the commands of rebalance process
func RebalanceCommand(c transaction.TxnCtx, cmd string) error {
	var rinfo RebalanceInfo
	err := c.Get("rinfo", &rinfo)
	if err != nil {
		return err
	}

	rebalanceProcess, err := NewRebalanceProcess(rinfo)
	if err != nil {
		return err
	}

	switch cmd {

	case "rebalanceStart":
		var r RebalanceInfo
		// Reset all variables
		glusterdvolinforesetstats(r)
		err = daemon.Start(rebalanceProcess, true)

	case "rebalanceStop":
		err = daemon.Stop(rebalanceProcess, true)

	case "rebalanceStatus":
		return nil
	}
	return err
}

// StartRebalance func used to start rebalance process
func StartRebalance(c transaction.TxnCtx) error {
	return RebalanceCommand(c, "rebalanceStart")
}

// StopRebalance func used to stop rebalance process
func StopRebalance(c transaction.TxnCtx) error {
	return RebalanceCommand(c, "rebalanceStop")
}

// StatusRebalance func used to get status of rebalance process
func StatusRebalance(c transaction.TxnCtx) error {
	return RebalanceCommand(c, "rebalanceStatus")
}
