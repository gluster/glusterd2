package device

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func deviceAddHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	peerID := mux.Vars(r)["peerid"]
	if uuid.Parse(peerID) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid peer-id passed in url")
		return
	}

	req := new(deviceapi.AddDeviceReq)
	if err := restutils.UnmarshalRequest(r, req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err)
		return
	}

	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Peer-id not found in store")
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, "Peer-id not found in store")
		return
	}

	devices, err := CheckIfDeviceExist(req.Devices, peerInfo.Metadata["_devices"])
	if err != nil {
		logger.WithError(err).WithField("device", req.Devices).Error("Device already exist")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Devices already exist")
		return
	}
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(peerID)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	txn.Nodes = []uuid.UUID{peerInfo.ID}
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "prepare-device",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	err = txn.Ctx.Set("peerid", &peerID)
	if err != nil {
		logger.WithError(err).WithField("key", "peerid").WithField("value", peerID).Error("Failed to set key in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	err = txn.Ctx.Set("devices", &devices)
	if err != nil {
		logger.WithError(err).WithField("key", "devices").Error("Failed to set key in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("Transaction to prepare device failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Transaction to prepare device failed")
		return
	}
	peerInfo, err = peer.GetPeer(peerID)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Failed to get peer from store")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to get peer from store")
		return
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, peerInfo)
}
