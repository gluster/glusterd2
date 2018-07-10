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

func validateSnapDeactivate(c transaction.TxnCtx) error {
	var brickinfos []brick.Brickinfo
	var snapname string

	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}

	vol := &snapinfo.SnapVolinfo
	if vol.State == volume.VolStarted {
		brickinfos, err = snapshot.GetOnlineBricks(vol)
		if err != nil {
			log.WithError(err).Error("failed to get online Bricks")
			return err
		}

	}
	if err := c.SetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	return nil
}

func deactivateSnapshot(c transaction.TxnCtx) error {
	var brickinfos []brick.Brickinfo
	var snapname string

	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}

	activate := false
	if err := c.GetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}
	//TODO Stop other process of snapshot volume
	//Yet to implement a generic way in glusterd2

	if err = snapshot.ActivateDeactivateFunc(snapinfo, brickinfos, activate, c.Logger()); err != nil {
		return err
	}
	vol := &snapinfo.SnapVolinfo
	for _, b := range vol.GetLocalBricks() {
		//Remove mount point of offline bricks if it present
		if snapshot.IsMountExist(b.Path, vol.ID) {
			snapshot.UmountBrick(b)
		}
	}

	return nil

}
func storeSnapshotDeactivate(c transaction.TxnCtx) error {
	var snapname string

	if err := c.Get("snapname", &snapname); err != nil {
		return err
	}

	snapinfo, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		return err
	}

	volinfo := &snapinfo.SnapVolinfo
	volinfo.State = volume.VolStopped

	if err := snapshot.AddOrUpdateSnapFunc(snapinfo); err != nil {
		c.Logger().WithError(err).WithField(
			"snapshot", volinfo.Name).Debug("storeSnapshot: failed to store snapshot info")
		return err
	}

	return nil
}

func rollbackDeactivateSnapshot(c transaction.TxnCtx) error {
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
	}
	if err = txn.Ctx.Set("snapname", &snapname); err != nil {
		log.WithError(err).Error("failed to set snap name in transaction context")
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
