package snapshotcommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/snapshot"
	"github.com/gluster/glusterd2/glusterd2/snapshot/lvm"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
)

const (
	brickStatusTxnKey string = "snapshotBrickstatuses"
)

func createSnapshotStatusResp(brickStatuses []brick.Brickstatus) []*api.SnapBrickStatus {
	var statusesRsp []*api.SnapBrickStatus
	for _, status := range brickStatuses {

		var s api.SnapBrickStatus
		s.Brick = api.BrickStatus{
			Info:      brick.CreateBrickInfo(&status.Info),
			Online:    status.Online,
			Pid:       status.Pid,
			Port:      status.Port,
			FS:        status.FS,
			MountOpts: status.MountOpts,
			Device:    status.Device,
			Size:      brick.CreateBrickSizeInfo(&status.Size),
		}

		if status.Online {
			lvs, err := lvm.GetLvsData(status.Device)
			if err == nil {
				s.LvData = lvm.CreateLvsResp(lvs)
			}
		}
		statusesRsp = append(statusesRsp, &s)
	}
	return statusesRsp

}

func snapshotStatus(ctx transaction.TxnCtx) error {
	var snapname string
	if err := ctx.Get("snapname", &snapname); err != nil {
		ctx.Logger().WithError(err).Error("Failed to get key from transaction context.")
		return err
	}

	snapshot, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to get snapshot information from store.")
		return err
	}
	vol := &snapshot.SnapVolinfo
	brickStatuses, err := volume.CheckBricksStatus(vol)
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to get brick status information.")
		return err
	}

	snapshotStatusesResp := createSnapshotStatusResp(brickStatuses)

	// Store the results in transaction context. This will be consumed by
	// the node that initiated the transaction.
	ctx.SetNodeResult(gdctx.MyUUID, brickStatusTxnKey, snapshotStatusesResp)
	return nil

}

func registerSnapshotStatusStepFuncs() {
	transaction.RegisterStepFunc(snapshotStatus, "snap-status.Check")
}

func snapshotStatusHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	snapname := mux.Vars(r)["snapname"]
	snap, err := snapshot.GetSnapshot(snapname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	vol := &snap.SnapVolinfo
	if vol.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrSnapNotActivated)
		return
	}
	txn, err := transaction.NewTxnWithLocks(ctx, vol.Name)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	defer txn.Done()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "snap-status.Check",
			Nodes:  vol.Nodes(),
		},
	}
	txn.Ctx.Set("snapname", snapname)

	// Some nodes may not be up, which is okay.
	txn.DontCheckAlive = true
	txn.DisableRollback = true

	err = txn.Do()
	if err != nil {
		logger.WithError(err).WithField("snapname", snapname).Error("Failed to get snapshot status")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	resp := createSnapshotStatusesResp(txn.Ctx, snap)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)
}

func createSnapshotStatusesResp(ctx transaction.TxnCtx, snap *snapshot.Snapinfo) *api.SnapStatusResp {

	// bmap is a map of brick statuses keyed by brick ID

	vol := snap.SnapVolinfo

	bmap := make(map[string]api.SnapBrickStatus)
	//Fill basic info of the bricks, in case a node is down, this data will be available
	for _, b := range vol.GetBricks() {
		bmap[b.ID.String()] = api.SnapBrickStatus{
			Brick: api.BrickStatus{
				Info: brick.CreateBrickInfo(&b),
			},
		}
	}

	// Loop over each node
	var resp api.SnapStatusResp
	resp.ParentName = snap.ParentVolume
	resp.SnapName = vol.Name
	resp.ID = vol.ID

	for _, node := range vol.Nodes() {
		var tmp []api.SnapBrickStatus
		err := ctx.GetNodeResult(node, brickStatusTxnKey, &tmp)
		if err != nil || len(tmp) == 0 {
			// skip if we do not have information
			continue
		}
		for _, b := range tmp {
			bmap[b.Brick.Info.ID.String()] = b
		}
	}

	for _, v := range bmap {
		resp.BrickStatus = append(resp.BrickStatus, v)
	}

	return &resp
}
