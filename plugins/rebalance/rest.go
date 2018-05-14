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
                Status:      rebalanceapi.Started,
		Cmd:         checkCmd(req),
		CommitHash:  setCommitHash(),
		RebalStats:       rebalanceapi.RebalNodeStatus{},
	}
}

func rebalanceStartHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// collect inputs from url
	volname := mux.Vars(r)["volname"]

	var req rebalanceapi.StartReq

//  TODO : How do I pass these options?
//  Unmarshal Request so to handle fix-layout and start force

//      if err := restutils.UnmarshalRequest(r, &req); err != nil {
//              restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
//              return
//	}

	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound)
		return
	}

	if vol.State != volume.VolStarted {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotStarted)
		return
	}

	if vol.DistCount == 1 {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrVolNotDistribute)
		return
	}

	// Check for remove- brick pending
	//TODO

	// A simple transaction to start rebalance on all nodes
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err)
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "rebalance-start",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "rebalance-store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).Error("failed to set volname in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	rebalinfo := createRebalanceInfo(volname, &req)
	if rebalinfo == nil {
		logger.WithError(err).Error("failed to create Rebalance info")
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
		 * rebalance process is one per node and depends on volfile change.
		 * Need to handle scenarios where process is started in
		 * few nodes and failed in few others */
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volname,
		}).Error("failed to start rebalance on volume")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	rebalinfo, err = GetRebalanceInfo(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
		return
	}

	logger.WithField("volname", rebalinfo.Volname).Info("rebalance started")

//TODO: Fix this!
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, rebalinfo)
}


func rebalanceStopHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	// collect inputs from url
	volname := mux.Vars(r)["volname"]

	// Validate rebalance command
	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound)
		return
	}

	rebalinfo, err := GetRebalanceInfo(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err)
		return
	}

	// Check rebalance process is started or not
	if rebalinfo.Status != rebalanceapi.Started {
		restutils.SendHTTPError(r.Context(), w, http.StatusBadRequest, errors.ErrRebalanceNotStarted)
		return
	}

	//A simple transaction to stop rebalance
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err.Error())
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "rebalance-stop",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "rebalance-store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		logger.WithError(err).Error("failed to set volname in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	rebalinfo.Volname = volname
	rebalinfo.Status = rebalanceapi.Stopped
	rebalinfo.Cmd = rebalanceapi.CmdStop

	err = txn.Ctx.Set("rinfo", rebalinfo)
	if err != nil {
		logger.WithError(err).Error("failed to set rebalance info in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volname,
		}).Error("failed to stop rebalance on volume")
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err.Error())
		return
	}

	txn.Ctx.Logger().WithField("volname", rebalinfo.Volname).Info("rebalance stopped")
	restutils.SendHTTPResponse(r.Context(), w, http.StatusOK, rebalinfo)
}




func rebalanceStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// collect inputs from url
	volname := mux.Vars(r)["volname"]

	// Validate rebalance command
	vol, err := volume.GetVolume(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrVolNotFound)
		return
	}

	rebalinfo, err := GetRebalanceInfo(volname)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, err.Error())
		return
	}


        // Get the consolidated status from all the nodes

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(volname)
	if err != nil {
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err.Error())
		return
	}

	err = txn.Ctx.Set("volname", volname)
	if err != nil {
		txn.Ctx.Logger().WithError(err).Error("failed to set volname in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	txn.Nodes = vol.Nodes()
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "rebalance-status",
			Nodes:  txn.Nodes,
		},
		{
			DoFunc: "rebalance-store",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
        }

	err = txn.Ctx.Set("rinfo", rebalinfo)
	if err != nil {
		txn.Ctx.Logger().WithError(err).Error("failed to set rebalance info in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error())
		return
	}

	err = txn.Do()
	if err != nil {
		txn.Ctx.Logger().WithFields(log.Fields{
			"error":   err.Error(),
			"volname": volname,
		}).Error("failed to stop rebalance on volume")
		restutils.SendHTTPError(r.Context(), w, http.StatusInternalServerError, err.Error())
		return
        }

        result, err := createRebalanceStatusResp(txn.Ctx, vol)
        if err != nil {
                errMsg := "Failed to aggregate rebalance status"
                txn.Ctx.Logger().WithField("error", err.Error()).Error("rebalanceStatusHandler:" + errMsg)
                restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, 
                        errMsg)
                return
        }

	restutils.SendHTTPResponse(r.Context(), w, http.StatusOK, result)
}



func createRebalanceStatusResp(ctx transaction.TxnCtx, volinfo *volume.Volinfo) (*rebalanceapi.RebalStatus, error) {
        var resp rebalanceapi.RebalStatus

        // Fill common info
        resp.Volname = volinfo.Name

        // Loop over each node of the volume and aggregate
        for _, node := range volinfo.Nodes() {
                var tmp rebalanceapi.RebalNodeStatus
                err := ctx.GetNodeResult(node, rebalStatusTxnKey, &tmp)
                if err != nil {
                        // skip if we do not have information
                        continue
                }

                resp.Nodes = append(resp.Nodes, tmp)
        }

        return &resp, nil
}
