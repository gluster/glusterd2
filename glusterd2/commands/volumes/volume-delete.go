package volumecommands

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	transactionv2 "github.com/gluster/glusterd2/glusterd2/transactionv2"
	"github.com/gluster/glusterd2/glusterd2/volume"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"

	"go.opencensus.io/trace"
)

func deleteVolume(c transaction.TxnCtx) error {

	var (
		volinfo volume.Volinfo
		err     error
	)
	if err = c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	err = volume.DeleteVolume(volinfo.Name)
	return err
}

func registerVolDeleteStepFuncs() {
	transaction.RegisterStepFunc(deleteVolume, "vol-delete.Store")
	transaction.RegisterStepFunc(txnCleanBricks, "vol-delete.CleanBricks")
}

func volumeDeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	volname := mux.Vars(r)["volname"]

	volinfo, status, err := DeleteVolume(ctx, volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	events.Broadcast(volume.NewEvent(volume.EventVolumeDeleted, volinfo))

	restutils.SendHTTPResponse(ctx, w, http.StatusNoContent, nil)
}

// DeleteVolume deletes the volume
func DeleteVolume(ctx context.Context, volname string) (*volume.Volinfo, int, error) {
	logger := gdctx.GetReqLogger(ctx)
	ctx, span := trace.StartSpan(ctx, "/volumeDeleteHandler")
	defer span.End()

	txn, err := transactionv2.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		return nil, status, err
	}
	defer txn.Done()

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		return nil, status, err
	}

	if volinfo.State == volume.VolStarted {
		return nil, http.StatusBadRequest, errors.New("volume must be in stopped state before deleting")
	}

	if len(volinfo.SnapList) > 0 {
		err = fmt.Errorf("cannot delete Volume %s ,as it has %d snapshots", volname, len(volinfo.SnapList))
		return nil, http.StatusFailedDependency, err
	}

	bricksAutoProvisioned := volinfo.IsAutoProvisioned() || volinfo.IsSnapshotProvisioned()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-delete.CleanBricks",
			Nodes:  volinfo.Nodes(),
			Skip:   !bricksAutoProvisioned,
		},
		{
			DoFunc: "vol-delete.Store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
			Sync:   true,
		},
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	span.AddAttributes(
		trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
		trace.StringAttribute("volName", volname),
	)

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("transaction to delete volume failed")
		return nil, http.StatusInternalServerError, err
	}

	return volinfo, http.StatusNoContent, nil
}
