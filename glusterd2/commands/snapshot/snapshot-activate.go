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
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func activateSnapshot(c transaction.TxnCtx) error {
	var snapinfo snapshot.Snapinfo
	activate := true

	if err := c.Get("oldsnapinfo", &snapinfo); err != nil {
		return err
	}
	vol := &snapinfo.SnapVolinfo

	brickinfos, err := snapshot.GetOfflineBricks(vol)
	if err != nil {
		c.Logger().WithError(err).Error("failed to get offline Bricks")
		return err
	}
	//Storing the data to use when rollbacking the operation
	if err := c.SetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	// Generate local Bricks Volfiles
	for _, b := range vol.GetLocalBricks() {
		volfileID := brick.GetVolfileID(vol.Name, b.Path)
		err := volgen.BrickVolfileToFile(vol, volfileID, "brick", b.PeerID.String(), b.Path)
		if err != nil {
			c.Logger().WithError(err).WithFields(log.Fields{
				"template": "brick",
				"volfile":  volfileID,
			}).Error("failed to generate volfile")
			return err
		}
	}

	err = snapshot.ActivateDeactivateFunc(&snapinfo, brickinfos, activate, c.Logger())
	return err
}

func rollbackActivateSnapshot(c transaction.TxnCtx) error {
	activate := false
	var snapinfo snapshot.Snapinfo
	var brickinfos []brick.Brickinfo

	if err := c.Get("oldsnapinfo", &snapinfo); err != nil {
		return err
	}
	vol := &snapinfo.SnapVolinfo

	if err := c.GetNodeResult(gdctx.MyUUID, "brickListToOperate", &brickinfos); err != nil {
		log.WithError(err).Error("failed to set request in transaction context")
		return err
	}

	// Remove the local Bricks Volfiles
	for _, b := range vol.GetLocalBricks() {
		volfileID := brick.GetVolfileID(vol.Name, b.Path)
		err := volgen.DeleteFile(volfileID)
		if err != nil {
			c.Logger().WithError(err).WithField("volfile", volfileID).Error("failed to delete volfile")
		}
	}

	err := snapshot.ActivateDeactivateFunc(&snapinfo, brickinfos, activate, c.Logger())

	return err

}

func registerSnapActivateStepFuncs() {
	transaction.RegisterStepFunc(activateSnapshot, "snap-activate.Commit")
	transaction.RegisterStepFunc(rollbackActivateSnapshot, "snap-activate.Undo")
	transaction.RegisterStepFunc(storeSnapshot, "snap-activate.StoreSnapshot")
	transaction.RegisterStepFunc(undoStoreSnapshot, "snap-activate.UndoStoreSnapshot")

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

	if vol.State == volume.VolStarted && req.Force == false {
		err := errors.New("snapshot already activated. Use force to override the behaviour")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	if err = txn.Ctx.Set("oldsnapinfo", &snapinfo); err != nil {
		log.WithError(err).Error("failed to set old snapinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	vol.State = volume.VolStarted
	if err = txn.Ctx.Set("snapinfo", &snapinfo); err != nil {
		log.WithError(err).Error("failed to set snapinfo in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	//Populating Nodes neeed not be under lock, because snapshot is a read only config
	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc:   "snap-activate.Commit",
			UndoFunc: "snap-activate.Undo",
			Nodes:    txn.Nodes,
		},
		{
			DoFunc:   "snap-activate.StoreSnapshot",
			UndoFunc: "snap-activate.UndoStoreSnapshot",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
		},
	}
	if err = txn.Do(); err != nil {
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

	resp := createSnapshotActivateResp(snapinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createSnapshotActivateResp(snap *snapshot.Snapinfo) *api.SnapshotActivateResp {
	return (*api.SnapshotActivateResp)(createSnapInfoResp(snap))
}
