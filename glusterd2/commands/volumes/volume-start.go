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
	var volname string
	if err := c.Get("volname", &volname); err != nil {
		return err
	}

	vol, err := volume.GetVolume(volname)
	if err != nil {
		return err
	}

	for _, b := range vol.GetLocalBricks() {

		c.Logger().WithFields(log.Fields{
			"volume": b.VolumeName,
			"brick":  b.String(),
		}).Info("Starting brick")

		if err := b.StartBrick(); err != nil {
			return err
		}
	}

	return nil
}

func stopAllBricks(c transaction.TxnCtx) error {
	var volname string
	if e := c.Get("volname", &volname); e != nil {
		c.Logger().WithFields(log.Fields{
			"error": e,
			"key":   "volname",
		}).Error("failed to get value for key from context")
		return e
	}

	vol, e := volume.GetVolume(volname)
	if e != nil {
		// this shouldn't happen
		c.Logger().WithFields(log.Fields{
			"error":   e,
			"volname": volname,
		}).Error("failed to get volinfo for volume")
		return e
	}

	for _, b := range vol.GetLocalBricks() {

		c.Logger().WithFields(log.Fields{
			"volume": b.VolumeName,
			"brick":  b.String(),
		}).Info("volume start failed, stopping brick")

		if err := b.StopBrick(); err != nil {
			return err
		}
	}

	return nil
}

func registerVolStartStepFuncs() {
	transaction.RegisterStepFunc(startAllBricks, "vol-start.Commit")
	transaction.RegisterStepFunc(stopAllBricks, "vol-start.Undo")
}

func volumeStartHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	volname := mux.Vars(r)["volname"]
	vol, e := volume.GetVolume(volname)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}
	if vol.State == volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolAlreadyStarted.Error(), api.ErrCodeDefault)
		return
	}

	// A simple one-step transaction to start the brick processes
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc:   "vol-start.Commit",
			UndoFunc: "vol-start.Undo",
			Nodes:    vol.Nodes(),
		},
		unlock,
	}
	txn.Ctx.Set("volname", volname)

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":  err.Error(),
			"volume": volname,
		}).Error("failed to start volume")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	vol.State = volume.VolStarted

	e = volume.AddOrUpdateVolumeFunc(vol)
	if e != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, e.Error(), api.ErrCodeDefault)
		return
	}

	events.Broadcast(newVolumeEvent(eventVolumeStarted, vol))
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, vol)
}
