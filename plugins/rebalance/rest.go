package rebalance

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/errors"

	rebalanceapi "github.com/gluster/glusterd2/plugins/rebalance/api"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func createRebalanceInfo(volname string, req *rebalanceapi.StartReq) *rebalanceapi.RebalInfo {
	return &rebalanceapi.RebalInfo{
		Volname:     volname,
		RebalanceID: uuid.NewRandom(),
		State:       rebalanceapi.Started,
		Cmd:         getCmd(req),
		CommitHash:  setCommitHash(),
		RebalStats:  []rebalanceapi.RebalNodeStatus{},
	}
}

func rebalanceStartHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// collect inputs from url
	volname := mux.Vars(r)["volname"]

	var req rebalanceapi.StartReq

	//  Unmarshal Request to check for fix-layout and start force
	err := restutils.UnmarshalRequest(r, &req)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	rebalinfo := createRebalanceInfo(volname, &req)
	if rebalinfo == nil {
		logger.WithError(err).Error("failed to create Rebalance info")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if rebalinfo.Cmd == rebalanceapi.CmdNone {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, ErrRebalanceInvalidOption)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	vol, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	if vol.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotStarted)
		return
	}

	if vol.DistCount == 1 {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, ErrVolNotDistribute)
		return
	}

	// TODO: Check for remove-brick

	// Start the rebalance process on all nodes
	// Only this node will save the rebalinfo in the store

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "rebalance-start",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "rebalance-store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).Error("failed to set volname in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Ctx.Set("rinfo", rebalinfo)
	if err != nil {
		logger.WithError(err).Error("failed to set rebalance info in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Do()
	if err != nil {
		/* TODO: Need to handle failure case. Unlike other process,
		 * rebalance process is one per node per volume.
		 * Need to handle scenarios where process is started in
		 * few nodes and failed in few others */
		logger.WithError(err).WithField("volname", volname).Error("failed to start rebalance on volume")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	rebalinfo, err = GetRebalanceInfo(volname)
	if err != nil {
		logger.WithError(err).WithField(
			"volname", volname).Error("failed to get the rebalance info for volume")
	}

	logger.WithField("volname", rebalinfo.Volname).Info("rebalance started")

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, rebalinfo.RebalanceID)
}

func rebalanceStopHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// collect inputs from url
	volname := mux.Vars(r)["volname"]

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	// Validate rebalance command
	vol, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	rebalinfo, err := GetRebalanceInfo(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, ErrRebalanceNotStarted)
		return
	}

	// Check whether the rebalance state is started
	if rebalinfo.State != rebalanceapi.Started {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, ErrRebalanceNotStarted)
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "rebalance-stop",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "rebalance-store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).Error("failed to set volname in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	rebalinfo.Volname = volname
	rebalinfo.State = rebalanceapi.Stopped
	rebalinfo.Cmd = rebalanceapi.CmdStop

	err = txn.Ctx.Set("rinfo", rebalinfo)
	if err != nil {
		logger.WithError(err).Error("failed to set rebalance info in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithError(err).WithField("volname", volname).Error("failed to stop rebalance on volume")
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err.Error())
		return
	}

	logger.WithField("volname", rebalinfo.Volname).Info("rebalance stopped")
	restutils.SendHTTPResponse(r.Context(), w, http.StatusOK, rebalinfo)
}

func rebalanceStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// collect inputs from url
	volname := mux.Vars(r)["volname"]

	txn, err := transaction.NewTxnWithLocks(ctx, volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	// Validate rebalance command
	vol, err := volume.GetVolume(volname)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	rebalinfo, err := GetRebalanceInfo(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err.Error())
		return
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).Error("failed to set volname in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	// The status will be a combination of those from the running rebalance processes
	// and the status stored in rebalinfo (by the processes that have completed)

	// Get the consolidated status from all the nodes
	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "rebalance-status",
			Nodes:  txn.Nodes,
		},
	}

	err = txn.Ctx.Set("rinfo", rebalinfo)
	if err != nil {
		logger.WithError(err).Error("failed to set rebalance info in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithError(err).WithField("volname", volname).Error("failed to query rebalance status for volume")
	}

	response, err := createRebalanceStatusResp(txn.Ctx, vol)
	if err != nil {
		errMsg := "Failed to create rebalance status response"
		logger.WithError(err).Error("rebalanceStatusHandler:" + errMsg)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError,
			errMsg)
		return
	}

	restutils.SendHTTPResponse(r.Context(), w, http.StatusOK, response)
}

func createRebalanceStatusResp(ctx transaction.TxnCtx, volinfo *volume.Volinfo) (*rebalanceapi.RebalStatus, error) {
	var (
		resp      rebalanceapi.RebalStatus
		tmp       rebalanceapi.RebalNodeStatus
		rebalinfo rebalanceapi.RebalInfo
	)

	err := ctx.Get("rinfo", &rebalinfo)
	if err != nil {
		log.WithField("volume", volinfo.Name).Error("Failed to get rebalinfo")
		return nil, err
	}

	// Fill common info
	resp.Volname = volinfo.Name
	resp.RebalanceID = rebalinfo.RebalanceID

	// Get the status for the completed processes first
	for _, tmp := range rebalinfo.RebalStats {
		resp.Nodes = append(resp.Nodes, tmp)
	}

	// Loop over each node of the volume and aggregate
	for _, node := range volinfo.Nodes() {
		err := ctx.GetNodeResult(node, rebalStatusTxnKey, &tmp)
		if err != nil {
			// skip. We might have it in the rebalinfo
			continue
		}

		resp.Nodes = append(resp.Nodes, tmp)
	}
	return &resp, nil
}
