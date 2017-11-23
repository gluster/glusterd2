package georeplication

import (
	"syscall"

	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"

	georepapi "github.com/gluster/glusterd2/plugins/georeplication/api"

	log "github.com/sirupsen/logrus"
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
