package device

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func deviceAddHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	req := new(deviceapi.AddDeviceReq)
	if err := restutils.UnmarshalRequest(r, req); err != nil {
		logger.WithError(err).Error("Failed to unmarshal request")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Unable to unmarshal request", api.ErrCodeDefault)
		return
	}
	peerID := mux.Vars(r)["peerid"]
	if uuid.Parse(peerID) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid peer id passed in request", api.ErrCodeDefault)
		return
	}
	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Peer ID not found in store")
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, "Peer Id not found in store", api.ErrCodeDefault)
		return
	}
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(peerInfo.ID.String())
	txn.Nodes = []uuid.UUID{peerInfo.ID}
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "prepare-device",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	err = txn.Ctx.Set("peerid", peerID)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Failed to set peerid in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	err = txn.Ctx.Set("req", req)
	if err != nil {
		logger.WithError(err).WithField("req-key", req).Error("Failed to set unmarshalled request information  in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("Transaction to prepare device failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Transaction to prepare device failed", api.ErrCodeDefault)
		return
	}
	peerInfo, err = peer.GetPeer(peerID)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Failed to get peer from store")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to get peer from store", api.ErrCodeDefault)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, peerInfo)
}

func peerEditGroupHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	req := new(deviceapi.PeerEditGroupReq)
	if err := restutils.UnmarshalRequest(r, req); err != nil {
		logger.WithError(err).Error("Failed to Unmarshal request")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Unable to unmarshal request", api.ErrCodeDefault)
		return
	}

	peerID := mux.Vars(r)["peerid"]
	if uuid.Parse(peerID) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid peer id passed in request", api.ErrCodeDefault)
		return
	}
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(peerID)
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "peer-edit-group",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}
	err = txn.Ctx.Set("peerid", peerID)
	if err != nil {
		logger.WithError(err).WithField("PeerID", peerID).Error("Failed to set peerid data in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	err = txn.Ctx.Set("req", req)
	if err != nil {
		logger.WithError(err).WithField("req", req).Error("Failed to set unmarshalled request data in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("Transaction to edit group failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Transaction to edit group failed", api.ErrCodeDefault)
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}
