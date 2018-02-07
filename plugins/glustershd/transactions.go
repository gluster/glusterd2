package glustershd

import (
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	volgen "github.com/gluster/glusterd2/glusterd2/volgen2"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"
)

type actionType uint16

const (
	actionStart actionType = iota
	actionStop
)

func selfhealdAction(c transaction.TxnCtx, action actionType) error {
	glustershDaemon, err := newGlustershd()
	if err != nil {
		return err
	}
	switch action {
	case actionStart:
		err = daemon.Start(glustershDaemon, true)
		if err == errors.ErrProcessAlreadyRunning {
			c.Logger().WithError(err).Error("Seld Heal Daemon is already running.")
			return nil
		}
	case actionStop:
		isVolRunning, err := volume.AreReplicateVolumesRunning()
		if err != nil {
			c.Logger().WithError(err).Error("Failed to get volinfo. Etcd server might be down.")
			return err
		}
		if !isVolRunning {
			err = daemon.Stop(glustershDaemon, true)
			if err != nil {
				c.Logger().WithError(err).Error("Failed to stop self heal daemon.")
				return err
			}
		}
	}
	return err
}

func txnSelfHealStart(c transaction.TxnCtx) error {
	return selfhealdAction(c, actionStart)
}

func txnSelfHealStop(c transaction.TxnCtx) error {
	return selfhealdAction(c, actionStop)
}

func txnSelfHealdUndo(c transaction.TxnCtx) error {
	var oldvolinfo volume.Volinfo
	if err := c.Get("oldvolinfo", &oldvolinfo); err != nil {
		return err
	}
	if err := volume.AddOrUpdateVolumeFunc(&oldvolinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", oldvolinfo.Name).Debug("storeVolume: failed to store volume info")
		return err
	}
	if err := volgen.Generate(); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", oldvolinfo.Name).Debug("generateVolfiles: failed to generate volfiles")
		return err
	}
	return nil
}
