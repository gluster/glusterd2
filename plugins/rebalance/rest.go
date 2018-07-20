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

	rebalInfo, err := GetRebalanceInfo(volname)
	if err == nil {
		if rebalInfo.State == rebalanceapi.Started {
			log.WithError(err).WithField("volume-name", volname).Error("Rebalance process has already been started.")
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest, ErrRebalanceAlreadyStarted)
			return
		}
	}

	rebalInfo = createRebalanceInfo(volname, &req)

	if rebalInfo.Cmd == rebalanceapi.CmdNone {
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
		logger.WithError(err).WithField("key", "volname").Error("failed to set volname in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Ctx.Set("rinfo", rebalInfo)
	if err != nil {
		logger.WithError(err).WithField("key", "rinfo").Error("failed to set rebalance info in transaction context")
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
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	logger.WithField("volname", rebalInfo.Volname).Info("rebalance started")
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
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

	rebalInfo, err := GetRebalanceInfo(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	// Check whether the rebalance state is started
	if rebalInfo.State != rebalanceapi.Started {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, ErrRebalanceNotStarted)
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
		logger.WithError(err).WithField("key", "volname").Error("failed to set volname in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	rebalInfo.State = rebalanceapi.Stopped
	rebalInfo.Cmd = rebalanceapi.CmdStop

	err = txn.Ctx.Set("rinfo", rebalInfo)
	if err != nil {
		logger.WithError(err).WithField("key", "rebalInfo").Error("failed to set rebalance info in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithError(err).WithField("volname", volname).Error("failed to stop rebalance on volume")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	logger.WithField("volname", rebalInfo.Volname).Info("rebalance stopped")
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
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

	rebalInfo, err := GetRebalanceInfo(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
		return
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).WithField("key", "volname").Error("failed to set volname in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
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

	err = txn.Ctx.Set("rinfo", rebalInfo)
	if err != nil {
		logger.WithError(err).WithField("key", "rinfo").Error("failed to set rebalance info in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithError(err).WithField("volname", volname).Error("failed to query rebalance status for volume")
		return
	}

	response, err := createRebalanceStatusResp(txn.Ctx, vol)
	if err != nil {
		errMsg := "failed to create rebalance status response"
		logger.WithError(err).Error("rebalanceStatusHandler:" + errMsg)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, errMsg)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, response)
}

func createRebalanceStatusResp(ctx transaction.TxnCtx, volInfo *volume.Volinfo) (*rebalanceapi.RebalStatus, error) {
	var (
		resp      rebalanceapi.RebalStatus
		tmp       rebalanceapi.RebalNodeStatus
		rebalInfo rebalanceapi.RebalInfo
	)

	err := ctx.Get("rinfo", &rebalInfo)
	if err != nil {
		log.WithError(err).WithField("volume", volInfo.Name).Error("Failed to get rebalance information")
		return nil, err
	}

	// Fill common info
	resp.Volname = volInfo.Name
	resp.RebalanceID = rebalInfo.RebalanceID
	resp.State = rebalInfo.State

	// Get the status for the completed processes first
	for _, tmp := range rebalInfo.RebalStats {
		resp.Nodes = append(resp.Nodes, tmp)
	}

	// Loop over each node of the volume and aggregate
	for _, node := range volInfo.Nodes() {
		err := ctx.GetNodeResult(node, rebalStatusTxnKey, &tmp)
		if err != nil {
			// skip. We might have it in the rebalinfo
			continue
		}

		resp.Nodes = append(resp.Nodes, tmp)
	}

	return &resp, nil
}
