package snapshotcommands

import (
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"

	log "github.com/sirupsen/logrus"
)

// undoStoreSnapshot revert back snapinfo and to generate client volfile
func undoStoreSnapshot(c transaction.TxnCtx) error {
	return storeSnapinfo(c, "oldsnapinfo")
}

// StoreSnapahot uses to store the snapinfo and to generate client volfile
func storeSnapshot(c transaction.TxnCtx) error {
	return storeSnapinfo(c, "snapinfo")
}

// storeSnapInfo uses to store the snapinfo based on key and to generate client volfile
func storeSnapinfo(c transaction.TxnCtx, key string) error {
	var snapinfo snapshot.Snapinfo

	if err := c.Get(key, &snapinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "snapinfo").Debug("Failed to get key from store")
		return err
	}
	volinfo := snapinfo.SnapVolinfo

	err := volgen.VolumeVolfileToStore(&volinfo, volinfo.Name, "client")
	if err != nil {
		c.Logger().WithError(err).WithFields(log.Fields{
			"template": "client",
			"volfile":  volinfo.Name,
		}).Error("failed to generate volfile and save to store")
		return err
	}

	if err := snapshot.AddOrUpdateSnapFunc(&snapinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"snapshot", volinfo.Name).Error("storeSnapshot: failed to store snapshot info")
		return err
	}

	/*
	   TODO
	   Intiate fetchspec notify to update snapd, once snapd is implemeted.
	*/

	return nil
}
