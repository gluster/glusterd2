package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func deleteVolume(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	if err := volume.DeleteVolume(volinfo.Name); err != nil {
		return err
	}

	if err := volgen.DeleteVolfiles(volinfo.Name); err != nil {
		c.Logger().WithError(err).
			WithField("volume", volinfo.Name).
			Warn("failed to delete volfiles of volume")
	}

	return nil
}

func registerVolDeleteStepFuncs() {
	transaction.RegisterStepFunc(deleteVolume, "vol-delete.Store")
	transaction.RegisterStepFunc(txnCleanBricks, "vol-delete.CleanBricks")
}

func volumeDeleteHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	if volinfo.State == volume.VolStarted {
		errMsg := "Volume must be in stopped state before deleting."
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
		return
	}

	bricksAutoProvisioned := false
	// TODO: Replace this with volinfo.Metadata["ProvisionState"] once available
	if len(volinfo.Subvols) > 0 && len(volinfo.Subvols[0].Bricks) > 0 {
		bricksAutoProvisioned = volinfo.Subvols[0].Bricks[0].DevicePath != ""
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-delete.CleanBricks",
			Nodes:  volinfo.Nodes(),
			Skip:   !bricksAutoProvisioned,
		},
		{
			DoFunc: "vol-delete.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("transaction to delete volume failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("volume-name", volinfo.Name).Info("volume deleted")
	events.Broadcast(volume.NewEvent(volume.EventVolumeDeleted, volinfo))

	restutils.SendHTTPResponse(ctx, w, http.StatusNoContent, nil)
}
