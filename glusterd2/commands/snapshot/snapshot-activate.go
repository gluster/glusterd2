package snapshotcommands

import (
	"errors"
	"io"
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
	var req api.VolCreateReq
	var brickinfos []brick.Brickinfo
	var snapname string

	if err := c.Get("req", &req); err != nil {
		return err
	}
	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}

	vol := &snapinfo.SnapVolinfo
	switch vol.State == volume.VolStarted {
	case true:
		if req.Force == false {
			return errors.New("snapshot already started. Use force to override the behaviour")
		}
		fallthrough
	case false:
		brickinfos, err = snapshot.GetOfflineBricks(vol)
		if err != nil {
			log.WithError(err).Error("failed to get offline Bricks")
			return err
		}
	}
	if err := c.SetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	return nil
}

func activateSnapshot(c transaction.TxnCtx) error {
	activate := true
	var brickinfos []brick.Brickinfo
	var snapname string

	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}
	if err := c.GetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	err = snapshot.ActivateDeactivateFunc(snapinfo, brickinfos, activate, c.Logger())
	if err != nil {
		return err
	}

	return nil

}
func storeSnapshotActivate(c transaction.TxnCtx) error {
	var snapname string

	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}

	volinfo := &snapinfo.SnapVolinfo
	volinfo.State = volume.VolStarted

	if err := snapshot.AddOrUpdateSnapFunc(snapinfo); err != nil {
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
	activate := false
	var brickinfos []brick.Brickinfo
	var snapname string

	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}

	if err := c.GetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	err = snapshot.ActivateDeactivateFunc(snapinfo, brickinfos, activate, c.Logger())

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

	txn, err := transaction.NewTxnWithLocks(ctx, snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	var req api.SnapActivateReq
	vol = &snapinfo.SnapVolinfo
	if err := restutils.UnmarshalRequest(r, &req); err != nil && err != io.EOF {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	err = txn.Ctx.Set("snapname", &snapname)
	if err != nil {
		log.WithError(err).Error("failed to set snap name in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	//Populating Nodes neeed not be under lock, because snapshot is a read only config
	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
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
	}
	err = txn.Ctx.Set("req", &req)
	if err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	err = txn.Do()
	if err != nil {
		log.WithError(err).WithField(
			"snapshot", snapname).Error("failed to start snapshot")
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
