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
	log "github.com/sirupsen/logrus"
)

func startAllBricks(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	for _, b := range volinfo.GetLocalBricks() {
		if err := StartBrick(b); err != nil {
			return err
		}
	}

	return nil
}

func stopAllBricks(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	for _, b := range volinfo.GetLocalBricks() {
		c.Logger().WithFields(log.Fields{
			"volume": b.VolumeName,
			"brick":  b.String(),
		}).Info("volume start failed, stopping brick")

		if err := b.StopBrickProcess(); err != nil {
			return err
		}
	}

	return nil
}

func registerVolStartStepFuncs() {
	transaction.RegisterStepFunc(startAllBricks, "vol-start.StartBricks")
	transaction.RegisterStepFunc(stopAllBricks, "vol-start.StartBricksUndo")
}

func volumeStartHandler(w http.ResponseWriter, r *http.Request) {

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

	if volinfo.State == volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolAlreadyStarted)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	txn.Steps = []*transaction.Step{
		{
			DoFunc:   "vol-start.StartBricks",
			UndoFunc: "vol-start.StartBricksUndo",
			Nodes:    volinfo.Nodes(),
		},
	}

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("transaction to start volume failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	volinfo.State = volume.VolStarted
	if err := volume.AddOrUpdateVolumeFunc(volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("volume-name", volinfo.Name).Info("volume started")
	events.Broadcast(newVolumeEvent(eventVolumeStarted, volinfo))

	resp := createVolumeStartResp(volinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeStartResp(v *volume.Volinfo) *api.VolumeStartResp {
	return (*api.VolumeStartResp)(volume.CreateVolumeInfoResp(v))
}
