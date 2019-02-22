package volumecommands

import (
	"net/http"
	"path/filepath"

	"github.com/gluster/glusterd2/glusterd2/events"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/oldtransaction"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/lvmutils"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	"go.opencensus.io/trace"
)

func registerVolExpandStepFuncs() {
	var sfs = []struct {
		name string
		sf   oldtransaction.StepFunc
	}{
		{"vol-expand.ValidateAndPrepare", expandValidatePrepare},
		{"vol-expand.ValidateBricks", validateBricks},
		{"vol-expand.InitBricks", initBricks},
		{"vol-expand.UndoInitBricks", undoInitBricks},
		{"vol-expand.GenerateBrickVolfiles", txnGenerateBrickVolfiles},
		{"vol-expand.GenerateBrickVolfiles.Undo", txnDeleteBrickVolfiles},
		{"vol-expand.StartBrick", startBricksOnExpand},
		{"vol-expand.UndoStartBrick", undoStartBricksOnExpand},
		{"vol-expand.UpdateVolinfo", updateVolinfoOnExpand},
		{"vol-expand.NotifyClients", notifyVolfileChange},
		{"vol-expand.LvmResize", resizeLVM},
	}
	for _, sf := range sfs {
		oldtransaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func validateVolumeExpandReq(req api.VolExpandReq) error {
	dupEntry := map[string]bool{}

	for _, brick := range req.Bricks {
		if dupEntry[brick.PeerID+filepath.Clean(brick.Path)] == true {
			return errors.ErrDuplicateBrickPath
		}
		dupEntry[brick.PeerID+filepath.Clean(brick.Path)] = true

	}

	return validateVolumeFlags(req.Flags)

}

// checkForLvmResize returns true if lvm resize is needed instead of adding new bricks or subvols
func checkForLvmResize(req api.VolExpandReq, volinfo *volume.Volinfo) bool {
	if req.DistributeCount == len(volinfo.Subvols) && req.Size != 0 {
		return true
	}
	return false
}

func volumeExpandHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ctx, span := trace.StartSpan(ctx, "/volumeExpandHandler")
	defer span.End()
	logger := gdctx.GetReqLogger(ctx)
	volname := mux.Vars(r)["volname"]

	var req api.VolExpandReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	if err := validateVolumeExpandReq(req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	txn, err := oldtransaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	volinfo, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	var expansionSizePerBrick uint64
	var expansionTpSizePerBrick uint64
	var expansionMetadataSizePerBrick uint64
	var brickVgMapping map[string]string
	var ok bool
	lvmResizeOp := checkForLvmResize(req, volinfo)
	// continue normal volume expand by adding new bricks or subvols
	if !lvmResizeOp {
		for index := range volinfo.Subvols {
			for _, brick := range volinfo.Subvols[index].Bricks {

				for _, b := range req.Bricks {

					if brick.PeerID.String() == b.PeerID && brick.Path == filepath.Clean(b.Path) {
						restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrDuplicateBrickPath)
						return
					}
				}

			}
		}
	} else {
		// lvmResize is needed
		switch volinfo.Type {
		case volume.Distribute:
			expansionSizePerSubvol := req.Size / uint64(len(volinfo.Subvols))
			expansionSizePerBrick = expansionSizePerSubvol
		case volume.Replicate, volume.DistReplicate:
			expansionSizePerSubvol := req.Size / uint64(len(volinfo.Subvols))
			expansionSizePerBrick = expansionSizePerSubvol
		case volume.Disperse, volume.DistDisperse:
			expansionSizePerSubvol := req.Size / uint64(len(volinfo.Subvols))
			expansionSizePerBrick = expansionSizePerSubvol / uint64(volinfo.Subvols[0].DisperseCount-volinfo.Subvols[0].RedundancyCount)
		}
		expansionTpSizePerBrick = uint64(float64(expansionSizePerBrick) * volinfo.SnapshotReserveFactor)
		expansionMetadataSizePerBrick = lvmutils.GetPoolMetadataSize(expansionTpSizePerBrick)
		totalExpansionSizePerBrick := expansionTpSizePerBrick + expansionMetadataSizePerBrick
		bricksInfo := volinfo.GetBricks()
		brickVgMapping, ok, err = deviceutils.CheckForAvailableVgSize(totalExpansionSizePerBrick, bricksInfo)
		if !ok && err == nil {
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Space not sufficient on device")
			return
		}

		if err != nil {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
			return
		}

	}

	nodes, err := req.Nodes()
	if err != nil {
		logger.WithError(err).Error("could not prepare node list")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	allNodes, err := peer.GetPeerIDs()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	txn.Nodes = allNodes
	txn.Steps = []*oldtransaction.Step{
		// TODO: This is a lot of steps. We can combine a few if we
		// do not re-use the same step functions across multiple
		// volume operations.
		{
			DoFunc: "vol-expand.ValidateAndPrepare",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
			Skip:   lvmResizeOp,
		},
		{
			DoFunc: "vol-expand.ValidateBricks",
			Nodes:  nodes,
			Skip:   lvmResizeOp,
			// Need to wait for newly selected bricks to be set by the previous step
			Sync: true,
		},
		{
			DoFunc:   "vol-expand.InitBricks",
			UndoFunc: "vol-expand.UndoInitBricks",
			Nodes:    nodes,
			Skip:     lvmResizeOp,
		},
		{
			DoFunc: "vol-expand.LvmResize",
			Nodes:  volinfo.Nodes(),
			Skip:   !lvmResizeOp,
		},
		{
			DoFunc:   "vol-create.StoreVolume",
			UndoFunc: "vol-create.UndoStoreVolume",
			Nodes:    []uuid.UUID{gdctx.MyUUID},
			Skip:     !lvmResizeOp,
			Sync:     true,
		},
		{
			DoFunc: "vol-expand.UpdateVolinfo",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
			Skip:   lvmResizeOp,
			Sync:   true,
		},
		{
			DoFunc:   "vol-expand.GenerateBrickVolfiles",
			UndoFunc: "vol-expand.GenerateBrickVolfiles.Undo",
			Nodes:    nodes,
			Skip:     (volinfo.State != volume.VolStarted),
		},
		{
			DoFunc:   "vol-expand.StartBrick",
			Nodes:    nodes,
			UndoFunc: "vol-expand.UndoStartBrick",
			Skip:     lvmResizeOp,
		},
		{
			DoFunc: "vol-expand.NotifyClients",
			Nodes:  allNodes,
			Skip:   lvmResizeOp,
		},
	}

	if err := txn.Ctx.Set("req", &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("volname", volname); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("expansionTpSizePerBrick", expansionTpSizePerBrick); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("expansionMetadataSizePerBrick", expansionMetadataSizePerBrick); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if err := txn.Ctx.Set("brickVgMapping", brickVgMapping); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	// Add relevant attributes to the root span
	var bricksToAdd string
	for _, b := range req.Bricks {
		bricksToAdd += b.PeerID + ":" + b.Path + ","
	}

	span.AddAttributes(
		trace.StringAttribute("reqID", txn.Ctx.GetTxnReqID()),
		trace.StringAttribute("volName", volname),
		trace.StringAttribute("bricksToAdd", bricksToAdd),
	)

	if err = txn.Do(); err != nil {
		logger.WithError(err).WithField("volume-name", volname).Error("volume expand transaction failed")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	volinfo, err = volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("volume-name", volinfo.Name).Info("volume expanded")
	events.Broadcast(volume.NewEvent(volume.EventVolumeExpanded, volinfo))

	resp := createVolumeExpandResp(volinfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createVolumeExpandResp(v *volume.Volinfo) *api.VolumeExpandResp {
	return (*api.VolumeExpandResp)(volume.CreateVolumeInfoResp(v))
}
