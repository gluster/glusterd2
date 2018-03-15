package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/cluster"
	"github.com/gluster/glusterd2/glusterd2/daemon"
	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/pmap"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"
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

		c.Logger().WithFields(log.Fields{
			"volume": b.VolumeName,
			"brick":  b.String(),
		}).Info("Starting brick")

		brickmux, err := cluster.IsBrickMuxEnabled()
		if err != nil {
			return err
		}

		if !brickmux {
			if err := b.StartBrick(); err != nil {
				return err
			}
			continue
		}

		// Find a compatible brick
		// If you can get the rpc client, then send an attach request

		compatBrick, err := FindCompatibleBrick(&b)
		if err != nil {
			return err
		}

		if compatBrick != nil {

			// Get the rpc client corresponding to the brick
			client, err := GetBrickRPCClient(compatBrick)
			if err != nil {
				// If you can't get an rpc client connection then start a separate brick process
				c.Logger().WithError(err).WithField(
					"brick", b.String()).Error("failed to connect to brick")
				if err := b.StartBrick(); err != nil {
					return err
				}
				return nil
			}

			req := &brick.GfBrickOpReq{
				Name: volgen.GetBrickVolFileID(b.VolumeName, b.NodeID.String(), b.Path),
				Op:   int(brick.OpBrickAttach),
			}

			var rsp brick.GfBrickOpRsp
			err = client.Call("Brick.OpBrickAttach", req, &rsp)
			if err != nil || rsp.OpRet != 0 {
				c.Logger().WithError(err).WithField(
					"brick", b.String()).Error("failed to send attach RPC, starting brick process")
				if err := b.StartBrick(); err != nil {
					return err
				}
			}

			compatPort := pmap.RegistrySearch(compatBrick.Path, pmap.GfPmapPortBrickserver)
			pmap.RegistryExtend(compatPort, b.Path, pmap.GfPmapPortBrickserver)

			brickDaemon, err := brick.NewGlusterfsd(b)
			if err != nil {
				return err
			}

			compatBrickDaemon, err := brick.NewGlusterfsd(*compatBrick)
			if err != nil {
				return err
			}

			if pidOnFile, err := daemon.ReadPidFromFile(compatBrickDaemon.PidFile()); err == nil {
				daemon.WritePidToFile(pidOnFile, brickDaemon.PidFile())
			}
		} else {
			if err := b.StartBrick(); err != nil {
				return err
			}
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

		if err := b.StopBrick(); err != nil {
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
