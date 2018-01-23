package snapshotcommands

import (
	"errors"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/plugins/snapshot"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func validateSnapDeactivate(c transaction.TxnCtx) error {
	var snapinfo snapshot.Snapinfo
	var brickinfos []brick.Brickinfo

	if err := c.Get("snapinfo", &snapinfo); err != nil {
		return err
	}
	vol := &snapinfo.SnapVolinfo
	switch vol.State == volume.VolStarted {
	case true:
		brickStatuses, err := volume.CheckBricksStatus(vol)
		if err != nil {
			return err
		}
		for _, brick := range brickStatuses {
			if brick.Online == true {
				brickinfos = append(brickinfos, brick.Info)
			}
		}
	case false:
		return errors.New("Snapshot is already stopped")

	}
	if err := c.SetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	return nil
}

func deactivateSnapshot(c transaction.TxnCtx) error {
	var snapinfo snapshot.Snapinfo
	var brickinfos []brick.Brickinfo
	activate := false
	if err := c.Get("snapinfo", &snapinfo); err != nil {
		return err
	}
	if err := c.GetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	err := snapshot.ActivateDeactivateFunc(&snapinfo, brickinfos, activate)
	if err != nil {
		return err
	}
	return nil

}
func storeSnapshotDeactivate(c transaction.TxnCtx) error {
	var snapInfo snapshot.Snapinfo
	if err := c.Get("snapinfo", &snapInfo); err != nil {
		return err
	}

	volinfo := &snapInfo.SnapVolinfo
	volinfo.State = volume.VolStopped

	if err := snapshot.AddOrUpdateSnapFunc(&snapInfo); err != nil {
		c.Logger().WithError(err).WithField(
			"snapshot", volinfo.Name).Debug("storeSnapshot: failed to store snapshot info")
		return err
	}

	return nil
}

func rollbackDeactivateSnapshot(c transaction.TxnCtx) error {
	var snapinfo snapshot.Snapinfo
	activate := true
	var brickinfos []brick.Brickinfo
	if err := c.Get("snapinfo", &snapinfo); err != nil {
		return err
	}

	if err := c.GetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	err := snapshot.ActivateDeactivateFunc(&snapinfo, brickinfos, activate)

	return err

}

func registerSnapDeactivateStepFuncs() {
	transaction.RegisterStepFunc(deactivateSnapshot, "snap-deactivate.Commit")
	transaction.RegisterStepFunc(rollbackDeactivateSnapshot, "snap-deactivate.Undo")
	transaction.RegisterStepFunc(storeSnapshotDeactivate, "snap-deactivate.StoreVolume")
	transaction.RegisterStepFunc(validateSnapDeactivate, "snap-deactivate.Validate")

}

func snapshotDeactivateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	var vol *volume.Volinfo

	snapname := mux.Vars(r)["snapname"]
	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
	}

	vol = &snapinfo.SnapVolinfo
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(snapname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "snap-deactivate.Validate",
			Nodes:  txn.Nodes,
		},

		{
			DoFunc:   "snap-deactivate.Commit",
			UndoFunc: "snap-deactivate.Undo",
			Nodes:    txn.Nodes,
		},
		{
			DoFunc: "snap-deactivate.StoreVolume",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},

		unlock,
	}
	err = txn.Ctx.Set("snapinfo", snapinfo)
	if err != nil {
		log.WithError(err).Error("failed to set snapinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Do()
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err.Error(),
			"snapshot": snapname,
		}).Error("failed to de-activate snap")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, vol)
}
