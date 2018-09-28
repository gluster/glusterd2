package volumecommands

import (
	"io"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

func startAllBricks(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	for _, b := range volinfo.GetLocalBricks() {

		c.Logger().WithFields(log.Fields{
			"volume": b.VolumeName,
			"brick":  b.String(),
		}).Info("Starting brick")

		if err := b.StartBrick(c.Logger()); err != nil {
			if err == errors.ErrProcessAlreadyRunning {
				continue
			}
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

		if err := b.StopBrick(c.Logger()); err != nil {
			return err
		}
	}

	return nil
}

func registerVolStartStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-start.StartBricks", startAllBricks},
		{"vol-start.StartBricksUndo", stopAllBricks},
		{"vol-start.XlatorActionDoVolumeStart", xlatorActionDoVolumeStart},
		{"vol-start.XlatorActionUndoVolumeStart", xlatorActionUndoVolumeStart},
		{"vol-start.UpdateVolinfo", storeVolume},
		{"vol-start.UpdateVolinfo.Undo", undoStoreVolume},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}

}

func volumeStartHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx, span := trace.StartSpan(ctx, "/volumeStartHandler")
	defer span.End()

	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]
	var req api.VolumeStartReq

	// request body is optional
	if err := restutils.UnmarshalRequest(r, &req); err != nil && err != io.EOF {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

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

	if volinfo.State == volume.VolStarted && !req.ForceStartBricks {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolAlreadyStarted)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc:   "vol-start.StartBricks",
			UndoFunc: "vol-start.StartBricksUndo",
			Nodes:    volinfo.Nodes(),
		},
		{
			DoFunc:   "vol-start.UpdateVolinfo",
			UndoFunc: "vol-start.UpdateVolinfo.Undo",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc:   "vol-start.XlatorActionDoVolumeStart",
			UndoFunc: "vol-start.XlatorActionUndoVolumeStart",
			Nodes:    volinfo.Nodes(),
		},
	}

	if err := txn.Ctx.Set("oldvolinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	volinfo.State = volume.VolStarted

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	span.AddAttributes(
		trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
		trace.StringAttribute("volName", volname),
	)

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("transaction to start volume failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("volume-name", volinfo.Name).Info("volume started")
	events.Broadcast(volume.NewEvent(volume.EventVolumeStarted, volinfo))

	resp := createVolumeStartResp(volinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeStartResp(v *volume.Volinfo) *api.VolumeStartResp {
	return (*api.VolumeStartResp)(volume.CreateVolumeInfoResp(v))
}
