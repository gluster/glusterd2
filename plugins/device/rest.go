package device

import (
	"encoding/json"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/errors"
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
		logger.WithError(err).Error("Failed to Unmarshal request")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	lock, unlock := transaction.CreateLockFuncs(peerID)
	if err := lock(ctx); err != nil {
		if err == transaction.ErrLockTimeout {
			restutils.SendHTTPError(ctx, w, http.StatusConflict, err)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		}
		return
	}
	defer unlock(ctx)

	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Peer ID not found in store")
		if err == errors.ErrPeerNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrPeerNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to get peer details from store")
		}
		return
	}

	var devices []deviceapi.Info
	err = json.Unmarshal([]byte(peerInfo.Metadata["_devices"]), &devices)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error(err)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	if checkIfDeviceExist(req.Device, devices) {
		logger.WithError(err).WithField("device", req.Device).Error("Device already exists")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Device already exists")
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	txn.Nodes = []uuid.UUID{peerInfo.ID}
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "prepare-device",
			Nodes:  txn.Nodes,
		},
	}
	err = txn.Ctx.Set("peerid", &peerID)
	if err != nil {
		logger.WithError(err).WithField("key", "peerid").WithField("value", peerID).Error("Failed to set key in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	err = txn.Ctx.Set("device", &req.Device)
	if err != nil {
		logger.WithError(err).WithField("key", "device").Error("Failed to set key in transaction context")
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
