package volumecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gorilla/mux"
)

const (
	brickStatusTxnKey string = "brickstatuses"
)

func registerBricksStatusStepFuncs() {
	transaction.RegisterStepFunc(bricksStatus, "bricks-status.Check")
}

func createBrickStatusRsp(brickStatuses []brick.Brickstatus) []*api.BrickStatus {
	var brickStatusesRsp []*api.BrickStatus
	for _, status := range brickStatuses {
		s := &api.BrickStatus{
			Info:      brick.CreateBrickInfo(&status.Info),
			Online:    status.Online,
			Pid:       status.Pid,
			Port:      status.Port,
			FS:        status.FS,
			MountOpts: status.MountOpts,
			Device:    status.Device,
			Size:      brick.CreateBrickSizeInfo(&status.Size),
		}
		brickStatusesRsp = append(brickStatusesRsp, s)
	}
	return brickStatusesRsp
}

func bricksStatus(ctx transaction.TxnCtx) error {
	var volname string
	if err := ctx.Get("volname", &volname); err != nil {
		ctx.Logger().WithError(err).Error("Failed to get key from transaction context.")
		return err
	}

	vol, err := volume.GetVolume(volname)
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to get volume information from store.")
		return err
	}
	brickStatuses, err := volume.CheckBricksStatus(vol)
	if err != nil {
		ctx.Logger().WithError(err).Error("Failed to get brick status information.")
		return err
	}
	brickStatusesRsp := createBrickStatusRsp(brickStatuses)
	// Store the results in transaction context. This will be consumed by
	// the node that initiated the transaction.
	ctx.SetNodeResult(gdctx.MyUUID, brickStatusTxnKey, brickStatusesRsp)
	return nil

}
func volumeBricksStatusHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	volname := mux.Vars(r)["volname"]
	vol, err := volume.GetVolume(volname)
	if err != nil {
		if err == errors.ErrVolNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "bricks-status.Check",
			Nodes:  vol.Nodes(),
		},
	}
	txn.Ctx.Set("volname", volname)

	// Some nodes may not be up, which is okay.
	txn.DontCheckAlive = true
	txn.DisableRollback = true

	err = txn.Do()
	if err != nil {
		logger.WithError(err).WithField("volume", volname).Error("Failed to get volume status")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	result, err := createBricksStatusResp(txn.Ctx, vol)
	if err != nil {
		errMsg := "Failed to aggregate brick status results from multiple nodes."
		logger.WithField("error", err.Error()).Error("volumeStatusHandler:" + errMsg)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, errMsg)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, result)
}

func createBricksStatusResp(ctx transaction.TxnCtx, vol *volume.Volinfo) (*api.BricksStatusResp, error) {

	// bmap is a map of brick statuses keyed by brick ID
	bmap := make(map[string]api.BrickStatus)

	// Loop over each node that make up the volume and aggregate result
	// of brick status check from each.
	var resp api.BricksStatusResp
	for _, node := range vol.Nodes() {
		var tmp []api.BrickStatus
		err := ctx.GetNodeResult(node, brickStatusTxnKey, &tmp)
		if err != nil || len(tmp) == 0 {
			// skip if we do not have information
			continue
		}
		for _, b := range tmp {
			bmap[b.Info.ID.String()] = b
		}
	}

	for _, v := range bmap {
		resp = append(resp, v)
	}

	return &resp, nil
}
