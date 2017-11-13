package georeplication

import (
	"github.com/gluster/glusterd2/bin/glusterd2/transaction"
	"github.com/gluster/glusterd2/daemon"
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

func startGsyncdMonitor(sess *georepapi.GeorepSession) error {
	gsyncdDaemon, err := newGsyncd(*sess)
	if err != nil {
		return err
	}

	err = daemon.Start(gsyncdDaemon, true)
	if err != nil {
		return err
	}
	return nil
}

func txnGeorepStart(c transaction.TxnCtx) error {
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
	}).Info("Starting gsyncd monitor")

	return startGsyncdMonitor(sessioninfo)
}
