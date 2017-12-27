package rebalance

import (
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/pborman/uuid"
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
		// Set rebalance status as started and store rebalance id
		var r RebalanceInfo
		r.Status = Started
		r.RebalanceID = uuid.NewRandom()
		if r.RebalanceID == nil {
			r.Status = Failed
			fmt.Println("Couldn't get rebalance id %s", r.Status)
			return nil
		}
		if r.Status != Started {
			fmt.Println("Couldn't update status %s", r.Status)
			return nil
		}
		// Reset all variables
		glusterdVolinfoResetStats(r)
		err = daemon.Start(rebalanceProcess, true)

	case "rebalanceStop":
		// Set rebalance status as stopped
		var r RebalanceInfo
		r.Status = Stopped
		if r.Status != Stopped {
			fmt.Println("Couldn't update status %s", r.Status)
			return nil
		}
		err = daemon.Stop(rebalanceProcess, true)

	}
	return err
}

// StartRebalance represents the rebalance start process
func StartRebalance(c transaction.TxnCtx) error {
	return RebalanceCommand(c, "rebalanceStart")
}

// StopRebalance represnts the rebalance stop process
func StopRebalance(c transaction.TxnCtx) error {
	return RebalanceCommand(c, "rebalanceStop")
}
