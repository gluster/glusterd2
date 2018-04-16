package snapshotcommands

import (
	"errors"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func validateSnapActivate(c transaction.TxnCtx) error {
	var snapinfo snapshot.Snapinfo
	var req api.VolCreateReq
	var brickinfos []brick.Brickinfo

	if err := c.Get("snapinfo", &snapinfo); err != nil {
		return err
	}

	if err := c.Get("req", &req); err != nil {
		return err
	}
	vol := &snapinfo.SnapVolinfo
	switch vol.State == volume.VolStarted {
	case true:
		if req.Force == false {
			return errors.New("Snapshot already started. Use force to override the behaviour")
		}
		fallthrough
	case false:
		brickStatuses, err := volume.CheckBricksStatus(vol)
		if err != nil {
			return err
		}

		for _, brick := range brickStatuses {
			if brick.Online == false {
				brickinfos = append(brickinfos, brick.Info)
			}
		}
	}
	if err := c.SetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	return nil
}

func activateSnapshot(c transaction.TxnCtx) error {
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
	if err != nil {
		return err
	}

	return nil

}
func storeSnapshotActivate(c transaction.TxnCtx) error {
	var snapInfo snapshot.Snapinfo
	if err := c.Get("snapinfo", &snapInfo); err != nil {
		return err
	}

	volinfo := &snapInfo.SnapVolinfo
	volinfo.State = volume.VolStarted

	if err := snapshot.AddOrUpdateSnapFunc(&snapInfo); err != nil {
		c.Logger().WithError(err).WithField(
			"snapshot", volinfo.Name).Debug("storeSnapshot: failed to store snapshot info")
		return err
	}
	/*
		TODO
		Intiate fetchspec notify to update snapd, once snapd is implemeted.
	*/

	return nil
}

func rollbackActivateSnapshot(c transaction.TxnCtx) error {
	var snapinfo snapshot.Snapinfo
	activate := false
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

func registerSnapActivateStepFuncs() {
	transaction.RegisterStepFunc(activateSnapshot, "snap-activate.Commit")
	transaction.RegisterStepFunc(rollbackActivateSnapshot, "snap-activate.Undo")
	transaction.RegisterStepFunc(storeSnapshotActivate, "snap-activate.StoreSnapshot")
	transaction.RegisterStepFunc(validateSnapActivate, "snap-activate.Validate")

}

func snapshotActivateHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	var vol *volume.Volinfo

	snapname := mux.Vars(r)["snapname"]
	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
	}

	var req api.SnapActivateReq
	vol = &snapinfo.SnapVolinfo
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, err)
		return
	}
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
			DoFunc: "snap-activate.Validate",
			Nodes:  txn.Nodes,
		},

		{
			DoFunc:   "snap-activate.Commit",
			UndoFunc: "snap-activate.Undo",
			Nodes:    txn.Nodes,
		},
		{
			DoFunc: "snap-activate.StoreSnapshot",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},

		unlock,
	}
	err = txn.Ctx.Set("req", &req)
	if err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
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
		}).Error("failed to start snapshot")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	snapinfo, err = snapshot.GetSnapshot(snapname)
	if err != nil {
		log.WithError(err).Error("failed to get snapinfo from store")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, snapinfo)
}
