package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/bin/glusterd2/gdctx"
	"github.com/gluster/glusterd2/bin/glusterd2/transaction"
	"github.com/gluster/glusterd2/brick"
	"github.com/gluster/glusterd2/daemon"
	"github.com/gluster/glusterd2/pkg/errors"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/volume"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func stopBricks(c transaction.TxnCtx) error {

	var volname string
	if err := c.Get("volname", &volname); err != nil {
		c.Logger().WithError(err).WithField(
			"key", "volname").Error("failed to get value for key from context")
		return err
	}

	vol, err := volume.GetVolume(volname)
	if err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volname).Error("failed to get volinfo for volume")
		return err
	}

	for _, b := range vol.Bricks {
		if uuid.Equal(b.NodeID, gdctx.MyUUID) {

			brickDaemon, err := brick.NewGlusterfsd(b)
			if err != nil {
				return err
			}

			c.Logger().WithFields(log.Fields{
				"volume": volname, "brick": b.String()}).Info("Stopping brick")

			client, err := daemon.GetRPCClient(brickDaemon)
			if err != nil {
				c.Logger().WithError(err).WithField(
					"brick", b.String()).Error("failed to connect to brick, sending SIGTERM")
				daemon.Stop(brickDaemon, false)
				continue
			}

			req := &brick.GfBrickOpReq{
				Name: b.Path,
				Op:   brick.OpBrickTerminate,
			}
			var rsp brick.GfBrickOpRsp
			err = client.Call("BrickOp", req, &rsp)
			if err != nil || rsp.OpRet != 0 {
				c.Logger().WithError(err).WithField(
					"brick", b.String()).Error("failed to send terminate RPC, sending SIGTERM")
				daemon.Stop(brickDaemon, false)
				continue
			}

			// On graceful shutdown of brick, daemon.Stop() isn't called.
			if err := daemon.DelDaemon(brickDaemon); err != nil {
				log.WithFields(log.Fields{
					"name": brickDaemon.Name(),
					"id":   brickDaemon.ID(),
				}).WithError(err).Warn("failed to delete brick entry from store, it may be restarted on GlusterD restart")
			}
		}
	}

	return nil
}

func registerVolStopStepFuncs() {
	transaction.RegisterStepFunc(stopBricks, "vol-stop.Commit")
}

func volumeStopHandler(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)
	volname := p["volname"]
	reqID, logger := restutils.GetReqIDandLogger(r)

	vol, e := volume.GetVolume(volname)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusNotFound, errors.ErrVolNotFound.Error())
		return
	}
	if vol.Status == volume.VolStopped {
		restutils.SendHTTPError(w, http.StatusBadRequest, errors.ErrVolAlreadyStopped.Error())
		return
	}

	// A simple one-step transaction to stop brick processes
	txn := transaction.NewTxn(reqID)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}
	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-stop.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("volname", volname)

	if _, err = txn.Do(); err != nil {
		logger.WithError(err).WithField(
			"volume", volname).Error("failed to stop volume")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(w, http.StatusConflict, err.Error())
		} else {
			restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	vol.Status = volume.VolStopped

	e = volume.AddOrUpdateVolumeFunc(vol)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	restutils.SendHTTPResponse(w, http.StatusOK, vol)
}
