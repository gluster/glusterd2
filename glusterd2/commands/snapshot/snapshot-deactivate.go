package snapshotcommands

import (
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

func deactivateSnapshot(c transaction.TxnCtx) error {
	var snapinfo snapshot.Snapinfo

	if err := c.Get("oldsnapinfo", &snapinfo); err != nil {
		return err
	}
	vol := &snapinfo.SnapVolinfo

	activate := false
	brickinfos, err := snapshot.GetOnlineBricks(vol)
	if err != nil {
		log.WithError(err).Error("failed to get online Bricks")
		return err
	}

	//Storing the value to do the rollback
	if err := c.SetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	//TODO Stop other process of snapshot volume
	//Yet to implement a generic way in glusterd2

	if err = snapshot.ActivateDeactivateFunc(&snapinfo, brickinfos, activate, c.Logger()); err != nil {
		return err
	}
	for _, b := range vol.GetLocalBricks() {
		//Remove mount point of offline bricks if it present
		if snapshot.IsMountExist(b.Path, vol.ID) {
			snapshot.UmountBrick(b)
		}
	}

	return nil

}

func rollbackDeactivateSnapshot(c transaction.TxnCtx) error {
	activate := true
	var snapinfo snapshot.Snapinfo
	var brickinfos []brick.Brickinfo

	if err := c.Get("oldsnapinfo", &snapinfo); err != nil {
		return err
	}

	if err := c.GetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	err := snapshot.ActivateDeactivateFunc(&snapinfo, brickinfos, activate, c.Logger())

	return err

}

func registerSnapDeactivateStepFuncs() {
	transaction.RegisterStepFunc(deactivateSnapshot, "snap-deactivate.Commit")
	transaction.RegisterStepFunc(rollbackDeactivateSnapshot, "snap-deactivate.Undo")
	transaction.RegisterStepFunc(storeSnapshot, "snap-deactivate.StoreSnapshot")
	transaction.RegisterStepFunc(undoStoreSnapshot, "snap-deactivate.UndoStoreSnapshot")

}

func snapshotDeactivateHandler(w http.ResponseWriter, r *http.Request) {

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

	vol = &snapinfo.SnapVolinfo
	if vol.State != volume.VolStarted {
		errMsg := "snapshot is already deactivated"
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc:   "snap-deactivate.Commit",
			UndoFunc: "snap-deactivate.Undo",
			Nodes:    txn.Nodes,
		},
		{
			DoFunc:   "snap-deactivate.StoreSnapshot",
			UndoFunc: "snap-deactivate.UndoStoreSnapshot",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
		},
	}
	if err = txn.Ctx.Set("oldsnapinfo", &snapinfo); err != nil {
		log.WithError(err).Error("failed to set old snapinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	vol.State = volume.VolStopped
	if err = txn.Ctx.Set("snapinfo", &snapinfo); err != nil {
		log.WithError(err).Error("failed to set snapinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err = txn.Do(); err != nil {
		log.WithError(err).WithField("snapshot", snapname).Error("failed to de-activate snap")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	//Fetching latest isnapinfo
	snapinfo, err = snapshot.GetSnapshot(snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	resp := createSnapshotDeactivateResp(snapinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createSnapshotDeactivateResp(snap *snapshot.Snapinfo) *api.SnapshotDeactivateResp {
	return (*api.SnapshotDeactivateResp)(createSnapInfoResp(snap))
}
