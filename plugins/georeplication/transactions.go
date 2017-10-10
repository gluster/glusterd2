package georeplication

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/gluster/glusterd2/daemon"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/transaction"
	"github.com/gluster/glusterd2/volume"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

const (
	gsyncdCmd                    = "/usr/local/libexec/glusterfs/gsyncd"
	gsyncdStartMaxRetries        = 10
	gsyncdStatusTxnKey    string = "gsyncdstatuses"
)

func txnGsyncdCreate(c transaction.TxnCtx) error {
	var sessioninfo Session
	if err := c.Get("geosession", &sessioninfo); err != nil {
		return err
	}

	if err := addOrUpdateSession(&sessioninfo); err != nil {
		c.Logger().WithError(err).WithField(
			"masterid", sessioninfo.MasterID).WithField(
			"slaveid", sessioninfo.SlaveID).Debug(
			"failed to store Geo-replication info")
		return err
	}

	return nil
}

func startGsyncdMonitor(sess *Session) error {

	gsyncdDaemon, err := NewGsyncd(*sess)
	if err != nil {
		return err
	}

	for i := 0; i < gsyncdStartMaxRetries; i++ {
		err = daemon.Start(gsyncdDaemon, true)
		if err != nil {
			return err
		}

		break
	}

	return nil
}

func stopGsyncdMonitor(sess *Session) error {
	gsyncdDaemon, err := NewGsyncd(*sess)
	if err != nil {
		return err
	}

	err = daemon.Stop(gsyncdDaemon, true)
	if err != nil {
		return err
	}

	return nil
}

func pauseGsyncdMonitor(sess *Session) error {
	gsyncdDaemon, err := NewGsyncd(*sess)
	if err != nil {
		return err
	}

	err = daemon.Pause(gsyncdDaemon)
	if err != nil {
		return err
	}

	return nil
}

func resumeGsyncdMonitor(sess *Session) error {
	gsyncdDaemon, err := NewGsyncd(*sess)
	if err != nil {
		return err
	}

	err = daemon.Resume(gsyncdDaemon)
	if err != nil {
		return err
	}

	return nil
}

func txnGsyncdStart(c transaction.TxnCtx) error {
	var masterid string
	var slaveid string
	if err := c.Get("masterid", &masterid); err != nil {
		return err
	}
	if err := c.Get("slaveid", &slaveid); err != nil {
		return err
	}

	sessioninfo, err := getSession(masterid, slaveid)
	if err != nil {
		return err
	}
	c.Logger().WithFields(log.Fields{
		"master": sessioninfo.MasterVol,
		"slave":  sessioninfo.SlaveHosts[0] + "::" + sessioninfo.SlaveVol,
	}).Info("Starting gsyncd monitor")

	if err := startGsyncdMonitor(sessioninfo); err != nil {
		return err
	}

	return nil
}

func txnGsyncdStop(c transaction.TxnCtx) error {
	var masterid string
	var slaveid string
	if err := c.Get("masterid", &masterid); err != nil {
		return err
	}
	if err := c.Get("slaveid", &slaveid); err != nil {
		return err
	}

	sessioninfo, err := getSession(masterid, slaveid)
	if err != nil {
		return err
	}
	c.Logger().WithFields(log.Fields{
		"master": sessioninfo.MasterVol,
		"slave":  sessioninfo.SlaveHosts[0] + "::" + sessioninfo.SlaveVol,
	}).Info("Stopping gsyncd monitor")

	if err := stopGsyncdMonitor(sessioninfo); err != nil {
		return err
	}

	return nil
}

func txnGsyncdPause(c transaction.TxnCtx) error {
	var masterid string
	var slaveid string
	if err := c.Get("masterid", &masterid); err != nil {
		return err
	}
	if err := c.Get("slaveid", &slaveid); err != nil {
		return err
	}

	sessioninfo, err := getSession(masterid, slaveid)
	if err != nil {
		return err
	}
	c.Logger().WithFields(log.Fields{
		"master": sessioninfo.MasterVol,
		"slave":  sessioninfo.SlaveHosts[0] + "::" + sessioninfo.SlaveVol,
	}).Info("Pausing gsyncd monitor")

	if err := pauseGsyncdMonitor(sessioninfo); err != nil {
		return err
	}

	return nil
}

func txnGsyncdResume(c transaction.TxnCtx) error {
	var masterid string
	var slaveid string
	if err := c.Get("masterid", &masterid); err != nil {
		return err
	}
	if err := c.Get("slaveid", &slaveid); err != nil {
		return err
	}

	sessioninfo, err := getSession(masterid, slaveid)
	if err != nil {
		return err
	}
	c.Logger().WithFields(log.Fields{
		"master": sessioninfo.MasterVol,
		"slave":  sessioninfo.SlaveHosts[0] + "::" + sessioninfo.SlaveVol,
	}).Info("Resuming gsyncd monitor")

	if err := resumeGsyncdMonitor(sessioninfo); err != nil {
		return err
	}

	return nil
}

func txnGsyncdStatus(c transaction.TxnCtx) error {
	var masterid string
	var slaveid string
	var err error

	if err = c.Get("masterid", &masterid); err != nil {
		return err
	}

	if err = c.Get("slaveid", &slaveid); err != nil {
		return err
	}

	sessioninfo, err := getSession(masterid, slaveid)
	if err != nil {
		return err
	}

	// Get Master vol info to get the bricks List
	volinfo, err := volume.GetVolume(sessioninfo.MasterVol)
	if err != nil {
		return err
	}

	var workersStatuses = make(map[string]Worker)

	for _, w := range volinfo.Bricks {

		if !uuid.Equal(w.NodeID, gdctx.MyUUID) {
			continue
		}

		// Example output: This may change in future, gsyncd can provide json
		// instead of this format.
		// checkpoint_time: N/A
		// last_synced_entry: 0
		// last_synced_utc: N/A
		// checkpoint_completion_time_utc: N/A
		// checkpoint_completed: N/A
		// meta: N/A
		// entry: N/A
		// slave_node: N/A
		// data: N/A
		// worker_status: Created
		// checkpoint_completion_time: N/A
		// checkpoint_completed_time: N/A
		// last_synced: N/A
		// checkpoint_time_utc: N/A
		// failures: N/A
		// crawl_status: N/A

		gsyncd, err := NewGsyncd(*sessioninfo)
		if err != nil {
			return err
		}
		args := gsyncd.StatusArgs(w.Path)

		out, err := exec.Command(gsyncdCmd, args).Output()
		if err != nil {
			return errors.New("Unable to execute gsyncd command")
		}

		data := strings.TrimSpace(string(out))

		for _, line := range strings.Split(data, "\n") {
			data := strings.SplitN(line, ":", 1)
			var worker Worker

			val := strings.TrimSpace(data[1])
			switch data[0] {
			case "checkpoint_time":
				worker.CheckpointTime = val
			case "last_synced_entry":
				worker.LastEntrySyncedTime = val
			case "last_synced_utc":
				worker.LastSyncedTimeUTC = val
			case "checkpoint_completion_time_utc":
				worker.CheckpointCompletedTimeUTC = val
			case "checkpoint_completed":
				if val == "Yes" {
					worker.CheckpointCompleted = true
				} else {
					worker.CheckpointCompleted = false
				}
			case "meta":
				worker.MetaOps = val
			case "entry":
				worker.EntryOps = val
			case "slave_node":
				worker.SlaveNode = val
			case "data":
				worker.DataOps = val
			case "worker_status":
				worker.Status = val
			case "checkpoint_completion_time":
				worker.CheckpointCompletedTime = val
			case "last_synced":
				worker.LastSyncedTime = val
			case "checkpoint_time_utc":
				worker.CheckpointTimeUTC = val
			case "failures":
				worker.FailedOps = val
			case "crawl_status":
				worker.CrawlStatus = val
			default:
			}

			// Unique key for master brick UUID:BRICK_PATH
			key := gdctx.MyUUID.String() + ":" + w.Path
			workersStatuses[key] = worker
		}
	}

	c.SetNodeResult(gdctx.MyUUID, gsyncdStatusTxnKey, workersStatuses)
	return nil
}

func aggregateGsyncdStatus(ctx transaction.TxnCtx, nodes []uuid.UUID) (*map[string]Worker, error) {
	var workersStatuses = make(map[string]Worker)

	// Loop over each node on which txn was run.
	// Fetch brick statuses stored by each node in transaction context.
	for _, node := range nodes {
		var tmp = make(map[string]Worker)
		err := ctx.GetNodeResult(node, gsyncdStatusTxnKey, &tmp)
		if err != nil {
			return nil, errors.New("aggregateGsyncdStatus: Could not fetch results from transaction context")
		}

		// Single final Hashmap
		for k, v := range tmp {
			workersStatuses[k] = v
		}
	}

	return &workersStatuses, nil
}
