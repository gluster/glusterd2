package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/daemon"
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
)

func stopBricks(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	for _, b := range volinfo.GetLocalBricks() {
		brickDaemon, err := brick.NewGlusterfsd(b)
		if err != nil {
			return err
		}

		c.Logger().WithFields(log.Fields{
			"volume": volinfo.Name, "brick": b.String()}).Info("Stopping brick")

		client, err := daemon.GetRPCClient(brickDaemon)
		if err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.String()).Error("failed to connect to brick, sending SIGTERM")
			daemon.Stop(brickDaemon, false, c.Logger())
			continue
		}

		req := &brick.GfBrickOpReq{
			Name: b.Path,
			Op:   int(brick.OpBrickTerminate),
		}
		var rsp brick.GfBrickOpRsp
		err = client.Call("Brick.OpBrickTerminate", req, &rsp)
		if err != nil || rsp.OpRet != 0 {
			c.Logger().WithError(err).WithField(
				"brick", b.String()).Error("failed to send terminate RPC, sending SIGTERM")
			daemon.Stop(brickDaemon, false, c.Logger())
			continue
		}

		// On graceful shutdown of brick, daemon.Stop() isn't called.
		if err := daemon.DelDaemon(brickDaemon); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"name": brickDaemon.Name(),
				"id":   brickDaemon.ID(),
			}).Warn("failed to delete brick entry from store, it may be restarted on GlusterD restart")
		}
	}

	return nil
}

func registerVolStopStepFuncs() {
	transaction.RegisterStepFunc(stopBricks, "vol-stop.StopBricks")
	transaction.RegisterStepFunc(storeVolume, "vol-stop.UpdateVolinfo")
	transaction.RegisterStepFunc(undoStoreVolume, "vol-stop.UpdateVolinfo.Undo")
}

func volumeStopHandler(w http.ResponseWriter, r *http.Request) {

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

	if volinfo.State == volume.VolStopped {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolAlreadyStopped)
		return
	}

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "vol-stop.StopBricks",
			Nodes:  volinfo.Nodes(),
		},
		{
			DoFunc:   "vol-stop.UpdateVolinfo",
			UndoFunc: "vol-stop.UpdateVolinfo.Undo",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
		},
	}

	if err := txn.Ctx.Set("oldvolinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	volinfo.State = volume.VolStopped

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

	logger.WithField("volume-name", volinfo.Name).Info("volume stopped")
	events.Broadcast(volume.NewEvent(volume.EventVolumeStopped, volinfo))

	resp := createVolumeStopResp(volinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeStopResp(v *volume.Volinfo) *api.VolumeStopResp {
	return (*api.VolumeStopResp)(volume.CreateVolumeInfoResp(v))
}
