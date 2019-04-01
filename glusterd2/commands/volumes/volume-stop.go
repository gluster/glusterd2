package volumecommands

import (
	"context"
	"net/http"
	"os"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/brickmux"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	transactionv2 "github.com/gluster/glusterd2/glusterd2/transactionv2"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

func stopBricks(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("volinfo", &volinfo); err != nil {
		return err
	}

	brickinfos := volinfo.GetLocalBricks()
	err := volgen.DeleteBricksVolfiles(brickinfos)
	if err != nil {
		return err
	}

	bmuxEnabled, err := brickmux.Enabled()
	if err != nil {
		return err
	}

	for _, b := range brickinfos {
		brickDaemon, err := brick.NewGlusterfsd(b)
		if err != nil {
			return err
		}

		if bmuxEnabled && !brickmux.IsLastBrickInProc(b) {
			c.Logger().WithFields(log.Fields{
				"volume": volinfo.Name, "brick": b.String()}).Info("Calling demultiplex for the brick")
			if err := brickmux.Demultiplex(b); err != nil {
				return err
			}
			c.Logger().WithFields(log.Fields{
				"volume": volinfo.Name, "brick": b.String()}).Info("deleting brick daemon from store")
			daemon.DelDaemon(brickDaemon)
			continue
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

		os.Remove(brickDaemon.PidFile())
		os.Remove(brickDaemon.SocketFile())
	}

	return nil
}

func registerVolStopStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-stop.StopBricks", stopBricks},
		{"vol-stop.XlatorActionDoVolumeStop", xlatorActionDoVolumeStop},
		{"vol-stop.XlatorActionUndoVOlumeStop", xlatorActionUndoVolumeStop},
		{"vol-stop.UpdateVolinfo", storeVolume},
		{"vol-stop.UpdateVolinfo.Undo", undoStoreVolume},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}

}

func volumeStopHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	volname := mux.Vars(r)["volname"]

	volinfo, status, err := StopVolume(ctx, volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	events.Broadcast(volume.NewEvent(volume.EventVolumeStopped, volinfo))
	resp := createVolumeStopResp(volinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

// StopVolume will stop the volume
func StopVolume(ctx context.Context, volname string) (*volume.Volinfo, int, error) {
	logger := gdctx.GetReqLogger(ctx)
	ctx, span := trace.StartSpan(ctx, "/volumeStopHandler")
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

	if volinfo.State == volume.VolStopped {
		return nil, http.StatusBadRequest, errors.ErrVolAlreadyStopped
	}

	if volinfo.State != volume.VolStarted {
		return nil, http.StatusBadRequest, errors.ErrVolNotStarted
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
			Sync:     true,
		},
		{
			DoFunc:   "vol-stop.XlatorActionDoVolumeStop",
			UndoFunc: "vol-stop.XlatorActionUndoVolumeStop",
			Nodes:    volinfo.Nodes(),
		},
	}

	if err := txn.Ctx.Set("oldvolinfo", volinfo); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	volinfo.State = volume.VolStopped

	if err := txn.Ctx.Set("volinfo", volinfo); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	span.AddAttributes(
		trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
		trace.StringAttribute("volName", volname),
	)

	if err := txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("transaction to stop volume failed")
		return nil, http.StatusInternalServerError, err
	}

	return volinfo, http.StatusOK, nil
}

func createVolumeStopResp(v *volume.Volinfo) *api.VolumeStopResp {
	return (*api.VolumeStopResp)(volume.CreateVolumeInfoResp(v))
}
