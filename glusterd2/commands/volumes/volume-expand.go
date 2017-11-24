package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volgen"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func checkBricksOnExpand(c transaction.TxnCtx) error {

	var newBricks []brick.Brickinfo
	if err := c.Get("newbricks", &newBricks); err != nil {
		return err
	}

	// TODO: Fix return values
	if _, err := volume.ValidateBrickEntriesFunc(newBricks, newBricks[0].VolumeID, true); err != nil {
		return err
	}

	return nil
}

func startBricksOnExpand(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("oldvolinfo", &volinfo); err != nil {
		return err
	}

	var newBricks []brick.Brickinfo
	if err := c.Get("newbricks", &newBricks); err != nil {
		return err
	}

	// Generate brick volfiles for the new bricks
	for _, b := range newBricks {
		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}
		if err := volgen.GenerateBrickVolfile(&volinfo, &b); err != nil {
			c.Logger().WithError(err).WithField(
				"brick", b.Path).Debug("GenerateBrickVolfile: failed to create brick volfile")
			return err
		}
	}

	if volinfo.State != volume.VolStarted {
		return nil
	}

	// Start the bricks
	for _, b := range newBricks {

		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}

		c.Logger().WithFields(log.Fields{
			"volume": b.VolumeName,
			"brick":  b.String(),
		}).Info("Starting brick")

		if err := startBrick(b); err != nil {
			return err
		}
	}

	return nil
}

func undoStartBricksOnExpand(c transaction.TxnCtx) error {

	var newBricks []brick.Brickinfo
	if err := c.Get("newbricks", &newBricks); err != nil {
		return err
	}

	// Stop the new bricks and delete brick volfile
	for _, b := range newBricks {

		if !uuid.Equal(b.NodeID, gdctx.MyUUID) {
			continue
		}

		c.Logger().WithFields(log.Fields{
			"volume": b.VolumeName,
			"brick":  b.String(),
		}).Info("volume expand failed, stopping brick")

		if err := stopBrick(b); err != nil {
			c.Logger().WithFields(log.Fields{
				"error":  err,
				"volume": b.VolumeName,
				"brick":  b.String(),
			}).Debug("stopping brick failed")
			// can't know here which of the new bricks started
			// so stopping brick might fail, but log anyway
		}

		if err := volgen.DeleteBrickVolfile(&b); err != nil {
			c.Logger().WithFields(log.Fields{
				"error":  err,
				"volume": b.VolumeName,
				"brick":  b.String(),
			}).Debug("failed to remove brick volfile")
		}
	}

	return nil
}

func updateVolinfoOnExpand(c transaction.TxnCtx) error {

	var newBricks []brick.Brickinfo
	if err := c.Get("newbricks", &newBricks); err != nil {
		return err
	}

	var volinfo volume.Volinfo
	if err := c.Get("oldvolinfo", &volinfo); err != nil {
		return err
	}

	var newReplicaCount int
	if err := c.Get("newreplicacount", &newReplicaCount); err != nil {
		return err
	}

	volinfo.ReplicaCount = newReplicaCount
	volinfo.Bricks = append(volinfo.Bricks, newBricks...)
	volinfo.DistCount = len(volinfo.Bricks) / volinfo.ReplicaCount

	switch len(volinfo.Bricks) {
	case volinfo.DistCount:
		volinfo.Type = volume.Distribute
	case volinfo.ReplicaCount:
		volinfo.Type = volume.Replicate
	default:
		volinfo.Type = volume.DistReplicate
	}

	// update new volinfo in txn ctx
	if err := c.Set("volinfo", volinfo); err != nil {
		return err
	}

	// update new volinfo in etcd store and generate client volfile
	if err := storeVolume(c); err != nil {
		c.Logger().WithError(err).WithField(
			"volume", volinfo.Name).Debug("storeVolume: failed to store volume info")
		return err
	}

	return nil
}

func registerVolExpandStepFuncs() {
	// NOTE: If txn steps are more granular, then the entire txn becomes more
	// resilient to recovery from failures i.e easier/better undo, but at
	// the expense of more number of co-ordinated network requests.
	transaction.RegisterStepFunc(checkBricksOnExpand, "vol-expand.CheckBrick")
	transaction.RegisterStepFunc(startBricksOnExpand, "vol-expand.StartBrick")
	transaction.RegisterStepFunc(undoStartBricksOnExpand, "vol-expand.UndoStartBrick")
	transaction.RegisterStepFunc(updateVolinfoOnExpand, "vol-expand.UpdateVolinfo") // only on initiator node
	transaction.RegisterStepFunc(notifyVolfileChange, "vol-expand.NotifyClients")
}

func volumeExpandHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	volname := mux.Vars(r)["volname"]

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound.Error(), api.ErrCodeDefault)
		return
	}

	var req api.VolExpandReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, errors.ErrJSONParsingFailed.Error(), api.ErrCodeDefault)
		return
	}

	newBrickCount := len(req.Bricks) + len(volinfo.Bricks)

	var newReplicaCount int
	if req.ReplicaCount != 0 {
		newReplicaCount = req.ReplicaCount
	} else {
		newReplicaCount = volinfo.ReplicaCount
	}

	if newBrickCount%newReplicaCount != 0 {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, "Invalid number of bricks", api.ErrCodeDefault)
		return
	}

	if volinfo.Type == volume.Replicate && req.ReplicaCount != 0 {
		if req.ReplicaCount < volinfo.ReplicaCount {
			restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, "Invalid number of bricks", api.ErrCodeDefault)
			return
		} else if req.ReplicaCount == volinfo.ReplicaCount {
			restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, "Replica count is same", api.ErrCodeDefault)
			return
		}
	}

	lock, unlock, err := transaction.CreateLockSteps(volinfo.Name)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	nodes, err := nodesFromBricks(req.Bricks)
	if err != nil {
		logger.WithError(err).Error("could not prepare node list")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Nodes = nodes
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-expand.CheckBrick",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc:   "vol-expand.StartBrick",
			Nodes:    txn.Nodes,
			UndoFunc: "vol-expand.UndoStartBrick",
		},
		{
			DoFunc: "vol-expand.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc: "vol-expand.NotifyClients",
			Nodes:  txn.Nodes,
		},
		unlock,
	}

	newBricks, err := volume.NewBrickEntriesFunc(req.Bricks, volinfo.Name, volinfo.ID)
	if err != nil {
		logger.WithError(err).Error("failed to create new brick entries")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err := txn.Ctx.Set("newbricks", newBricks); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err := txn.Ctx.Set("newreplicacount", newReplicaCount); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err := txn.Ctx.Set("oldvolinfo", volinfo); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if _, err = txn.Do(); err != nil {
		logger.WithError(err).Error("volume expand transaction failed")
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err.Error(), api.ErrCodeDefault)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		}
		return
	}

	newvolinfo, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	resp := createVolumeExpandResp(newvolinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeExpandResp(v *volume.Volinfo) *api.VolumeExpandResp {
	return (*api.VolumeExpandResp)(createVolumeInfoResp(v))
}
