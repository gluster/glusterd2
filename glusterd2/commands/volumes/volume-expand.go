package volumecommands

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
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

func startBricksOnExpand(c transaction.TxnCtx) error {

	var volinfo volume.Volinfo
	if err := c.Get("oldvolinfo", &volinfo); err != nil {
		return err
	}

	var newBricks []brick.Brickinfo
	if err := c.Get("bricks", &newBricks); err != nil {
		return err
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

		if err := b.StartBrick(); err != nil {
			return err
		}
	}

	return nil
}

func undoStartBricksOnExpand(c transaction.TxnCtx) error {

	var newBricks []brick.Brickinfo
	if err := c.Get("bricks", &newBricks); err != nil {
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

		if err := b.StopBrick(); err != nil {
			c.Logger().WithFields(log.Fields{
				"error":  err,
				"volume": b.VolumeName,
				"brick":  b.String(),
			}).Debug("stopping brick failed")
			// can't know here which of the new bricks started
			// so stopping brick might fail, but log anyway
		}

	}

	return nil
}

func updateVolinfoOnExpand(c transaction.TxnCtx) error {

	var newBricks []brick.Brickinfo
	if err := c.Get("bricks", &newBricks); err != nil {
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

	// TODO: Assumption, all subvols are same
	// If New Replica count is different than existing then add one brick to each subvolume
	// Or if the Volume consists of only one subvolume.
	var addNewSubvolume bool
	switch volinfo.Subvols[0].Type {
	case volume.SubvolDistribute:
		addNewSubvolume = false
	case volume.SubvolReplicate:
		if newReplicaCount != volinfo.Subvols[0].ReplicaCount {
			addNewSubvolume = false
		}
	default:
		addNewSubvolume = true
	}

	if !addNewSubvolume {
		idx := 0
		for _, b := range newBricks {
			// If number of bricks specified in add brick is more than
			// the number of sub volumes. For example, if number of subvolumes is 2
			// but 4 bricks specified in add brick command.
			if idx >= len(volinfo.Subvols) {
				idx = 0
			}
			volinfo.Subvols[idx].Bricks = append(volinfo.Subvols[idx].Bricks, b)
		}
	} else {
		// Create new Sub volumes with given bricks
		subvolIdx := len(volinfo.Subvols)
		for i := 0; i < len(newBricks)/newReplicaCount; i++ {
			idx := i * newReplicaCount
			volinfo.Subvols = append(volinfo.Subvols, volume.Subvol{
				ID:     uuid.NewRandom(),
				Name:   fmt.Sprintf("%s-%s-%d", volinfo.Name, strings.ToLower(volinfo.Subvols[0].Type.String()), subvolIdx),
				Type:   volinfo.Subvols[0].Type,
				Bricks: newBricks[idx : idx+newReplicaCount],
			})
			subvolIdx = subvolIdx + 1
		}
	}

	// Update all Subvols Replica count
	for idx := range volinfo.Subvols {
		volinfo.Subvols[idx].ReplicaCount = newReplicaCount
	}

	volinfo.DistCount = len(volinfo.Subvols)

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
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"vol-expand.ValidateBricks", validateBricks},
		{"vol-expand.InitBricks", initBricks},
		{"vol-expand.UndoInitBricks", undoInitBricks},
		{"vol-expand.StartBrick", startBricksOnExpand},
		{"vol-expand.UndoStartBrick", undoStartBricksOnExpand},
		{"vol-expand.UpdateVolinfo", updateVolinfoOnExpand},
		{"vol-expand.NotifyClients", notifyVolfileChange},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
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

	numBricks := 0
	for _, subvol := range volinfo.Subvols {
		numBricks += len(subvol.Bricks)
	}
	newBrickCount := len(req.Bricks) + numBricks

	var newReplicaCount int
	if req.ReplicaCount != 0 {
		newReplicaCount = req.ReplicaCount
	} else {
		newReplicaCount = volinfo.Subvols[0].ReplicaCount
	}

	if newBrickCount%newReplicaCount != 0 {
		restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, "Invalid number of bricks", api.ErrCodeDefault)
		return
	}

	if volinfo.Type == volume.Replicate && req.ReplicaCount != 0 {
		// TODO: Only considered first sub volume's ReplicaCount
		if req.ReplicaCount < volinfo.Subvols[0].ReplicaCount {
			restutils.SendHTTPError(ctx, w, http.StatusUnprocessableEntity, "Invalid number of bricks", api.ErrCodeDefault)
			return
		} else if req.ReplicaCount == volinfo.Subvols[0].ReplicaCount {
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

	nodes, err := nodesFromVolumeExpandReq(&req)
	if err != nil {
		logger.WithError(err).Error("could not prepare node list")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "vol-expand.ValidateBricks",
			Nodes:  nodes,
		},
		{
			DoFunc:   "vol-expand.InitBricks",
			UndoFunc: "vol-expand.UndoInitBricks",
			Nodes:    nodes,
		},
		{
			DoFunc: "vol-expand.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		{
			DoFunc:   "vol-expand.StartBrick",
			Nodes:    nodes,
			UndoFunc: "vol-expand.UndoStartBrick",
		},
		{
			DoFunc: "vol-expand.NotifyClients",
			Nodes:  nodes,
		},
		unlock,
	}

	newBricks, err := volume.NewBrickEntriesFunc(req.Bricks, volinfo.Name, volinfo.ID)
	if err != nil {
		logger.WithError(err).Error("failed to create new brick entries")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	if err := txn.Ctx.Set("bricks", newBricks); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	var checks brick.InitChecks
	if !req.Force {
		checks.IsInUse = true
		checks.IsMount = true
		checks.IsOnRoot = true
	}

	err = txn.Ctx.Set("brick-checks", &checks)
	if err != nil {
		logger.WithError(err).WithField("key", "brick-checks").Error("failed to set key in transaction context")
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

	if err = txn.Do(); err != nil {
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
	return (*api.VolumeExpandResp)(volume.CreateVolumeInfoResp(v))
}
