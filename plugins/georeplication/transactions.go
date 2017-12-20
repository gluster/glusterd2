package georeplication

import (
	"encoding/json"
	"errors"
	"os/exec"
	"syscall"

	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"

	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	gsyncdStatusTxnKey string = "gsyncdstatuses"
)

func txnGeorepCreate(c transaction.TxnCtx) error {
	var sessioninfo georepapi.GeorepSession
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

func gsyncdAction(c transaction.TxnCtx, action actionType) error {
	var masterid string
	var slaveid string
	if err := c.Get("mastervolid", &masterid); err != nil {
		return err
	}
	if err := c.Get("slavevolid", &slaveid); err != nil {
		return err
	}

	sessioninfo, err := getSession(masterid, slaveid)
	if err != nil {
		return err
	}
	c.Logger().WithFields(log.Fields{
		"master": sessioninfo.MasterVol,
		"slave":  sessioninfo.SlaveHosts[0] + "::" + sessioninfo.SlaveVol,
	}).Info(action.String() + " gsyncd monitor")

	gsyncdDaemon, err := newGsyncd(*sessioninfo)
	if err != nil {
		return err
	}

	switch action {
	case actionStart:
		err = daemon.Start(gsyncdDaemon, true)
	case actionStop:
		err = daemon.Stop(gsyncdDaemon, true)
	case actionPause:
		err = daemon.Signal(gsyncdDaemon, syscall.SIGSTOP)
	case actionResume:
		err = daemon.Signal(gsyncdDaemon, syscall.SIGCONT)
	}

	return err
}

func txnGeorepStart(c transaction.TxnCtx) error {
	return gsyncdAction(c, actionStart)
}

func txnGeorepStop(c transaction.TxnCtx) error {
	return gsyncdAction(c, actionStop)
}

func txnGeorepDelete(c transaction.TxnCtx) error {
	var masterid string
	var slaveid string
	if err := c.Get("mastervolid", &masterid); err != nil {
		return err
	}
	if err := c.Get("slavevolid", &slaveid); err != nil {
		return err
	}

	sessioninfo, err := getSession(masterid, slaveid)
	if err != nil {
		return err
	}

	if err := deleteSession(masterid, slaveid); err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"master": sessioninfo.MasterVol,
			"slave":  sessioninfo.SlaveHosts[0] + "::" + sessioninfo.SlaveVol,
		}).Debug("failed to delete Geo-replication info from store")
		return err
	}

	return nil
}

func txnGeorepPause(c transaction.TxnCtx) error {
	return gsyncdAction(c, actionPause)
}

func txnGeorepResume(c transaction.TxnCtx) error {
	return gsyncdAction(c, actionResume)
}

func txnGeorepStatus(c transaction.TxnCtx) error {
	var masterid string
	var slaveid string
	var err error

	if err = c.Get("mastervolid", &masterid); err != nil {
		return err
	}

	if err = c.Get("slavevolid", &slaveid); err != nil {
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

	var workersStatuses = make(map[string]georepapi.GeorepWorker)

	for _, subvol := range volinfo.Subvols {
		for _, w := range subvol.Bricks {

			if !uuid.Equal(w.NodeID, gdctx.MyUUID) {
				continue
			}

			gsyncd, err := newGsyncd(*sessioninfo)
			if err != nil {
				return err
			}
			args := gsyncd.statusArgs(w.Path)

			out, err := exec.Command(gsyncdCommand, args...).Output()
			if err != nil {
				return err
			}

			var worker georepapi.GeorepWorker
			if err = json.Unmarshal(out, &worker); err != nil {
				return err
			}

			// Unique key for master brick UUID:BRICK_PATH
			key := gdctx.MyUUID.String() + ":" + w.Path
			workersStatuses[key] = worker
		}
	}

	c.SetNodeResult(gdctx.MyUUID, gsyncdStatusTxnKey, workersStatuses)
	return nil
}

func aggregateGsyncdStatus(ctx transaction.TxnCtx, nodes []uuid.UUID) (*map[string]georepapi.GeorepWorker, error) {
	var workersStatuses = make(map[string]georepapi.GeorepWorker)

	// Loop over each node on which txn was run.
	// Fetch brick statuses stored by each node in transaction context.
	for _, node := range nodes {
		var tmp = make(map[string]georepapi.GeorepWorker)
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
