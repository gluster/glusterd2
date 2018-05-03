package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
)

func stopBricks(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	for _, b := range volinfo.GetLocalBricks() {
		if err := b.StopBrickProcess(); err != nil {
			return err
		}
		continue
	}

	return nil
}

func registerVolStopStepFuncs() {
	transaction.RegisterStepFunc(stopBricks, "vol-stop.StopBricks")
}

func volumeStopHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]

	lock, unlock := transaction.CreateLockFuncs(volname)
	// Taking a lock outside the txn as volinfo.Nodes() must also
	// be populated holding the lock. See issue #510
	if err := lock(ctx); err != nil {
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}
	defer unlock(ctx)

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		if err == errors.ErrVolNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}

	if volinfo.State == volume.VolStopped {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolAlreadyStopped)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-stop.StopBricks",
			Nodes:  volinfo.Nodes(),
		},
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("transaction to stop volume failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	volinfo.State = volume.VolStopped
	if err := volume.AddOrUpdateVolumeFunc(volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("volume-name", volinfo.Name).Info("volume stopped")
	events.Broadcast(newVolumeEvent(eventVolumeStopped, volinfo))

	resp := createVolumeStopResp(volinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeStopResp(v *volume.Volinfo) *api.VolumeStopResp {
	return (*api.VolumeStopResp)(volume.CreateVolumeInfoResp(v))
}
