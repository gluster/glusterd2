package rebalance

import (
	"context"
	"errors"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"

	log "github.com/sirupsen/logrus"
)

//HandleEventNotify updates the rebalinfo in the store with the status sent by
// the rebalance process
func HandleEventNotify(status map[string]string) error {
	var (
		ok      bool
		volname string
		err     error
	)

	volname, ok = status["volname"]
	if !ok {
		err = errors.New("volname key not found")
		return err
	}

	volname = strings.TrimLeft(volname, "rebalance/")
	log.Debug("In RebalanceHandleEventNotify " + volname)

	var rebalinfo *rebalanceapi.RebalInfo
	var rebalNodeStatus rebalanceapi.RebalNodeStatus

	txn, err := transaction.NewTxnWithLocks(context.TODO(), volname)
	if err != nil {
		log.WithError(err).Error("Locking failed. Unable to update store")
		return err
	}
	defer txn.Done()

	vol, err := volume.GetVolume(volname)
	if err != nil {
		return err
	}

	rebalinfo, err = GetRebalanceInfo(volname)
	if err != nil {
		log.WithError(err).Error("Failed to get rebalance info from store")
		return err
	}

	rebalNodeStatus.PeerID = gdctx.MyUUID
	rebalNodeStatus.Status = status["status"]
	rebalNodeStatus.RebalancedFiles = status["files"]
	rebalNodeStatus.RebalancedSize = status["size"]
	rebalNodeStatus.LookedupFiles = status["lookups"]
	rebalNodeStatus.SkippedFiles = status["skipped"]
	rebalNodeStatus.RebalanceFailures = status["failures"]
	rebalNodeStatus.ElapsedTime = status["run-time"]
	rebalNodeStatus.TimeLeft = status["time-left"]

	rebalinfo.RebalStats = append(rebalinfo.RebalStats, rebalNodeStatus)
	if len(rebalinfo.RebalStats) == len(vol.Nodes()) {
		rebalinfo.State = rebalanceapi.Complete
	}

	err = StoreRebalanceInfo(rebalinfo)
	if err != nil {
		log.WithError(err).Error("Failed to store rebalance info")
		return err
	}

	return nil
}
