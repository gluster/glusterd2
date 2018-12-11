package device

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/errors"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func deviceAddHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	peerID := mux.Vars(r)["peerid"]
	if uuid.Parse(peerID) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "invalid peer-id passed in url")
		return
	}

	req := new(deviceapi.AddDeviceReq)
	if err := restutils.UnmarshalRequest(r, req); err != nil {
		logger.WithError(err).Error("Failed to unmarshal request")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, peerID)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	_, err = deviceutils.GetDevice(peerID, req.Device)
	if err == nil {
		logger.WithError(err).WithField("device", req.Device).Error("Device already exists")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "device already exists")
		return
	}

	if err != errors.ErrDeviceNotFound {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Peer ID not found in store")
		if err == errors.ErrPeerNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrPeerNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "failed to get peer details from store")
		}
		return
	}

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
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "transaction to prepare device failed")
		return
	}
	deviceInfo, err := deviceutils.GetDevice(peerID, req.Device)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Failed to get device from store")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "failed to get device from store")
		return
	}

	restutils.SetLocationHeader(r, w, strings.TrimLeft(deviceInfo.Device, "/"))
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, deviceInfo)
}

func deviceListHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	peerID := mux.Vars(r)["peerid"]
	if peerID != "" && uuid.Parse(peerID) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "invalid peer-id passed in url")
		return
	}

	devName := mux.Vars(r)["device"]
	devName = "/" + devName

	var err error
	var devices []deviceapi.Info
	if peerID == "" {
		devices, err = deviceutils.GetDevices()
	} else {
		if devName != "/" {
			var dev *deviceapi.Info
			dev, err = deviceutils.GetDevice(peerID, devName)
			if err == nil {
				devices = []deviceapi.Info{*dev}
			}
		} else {
			devices, err = deviceutils.GetDevices(peerID)
		}
	}
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error(
			"Failed to get devices for peer")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, devices)
}

func deviceEditHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	peerID := mux.Vars(r)["peerid"]
	if uuid.Parse(peerID) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "invalid peer-id passed in url")
		return
	}

	device := mux.Vars(r)["device"]
	if device == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "device not provided in URL")
		return
	}

	// Adding prefix (/) to device
	device = "/" + device

	req := new(deviceapi.EditDeviceReq)
	if err := restutils.UnmarshalRequest(r, req); err != nil {
		logger.WithError(err).Error("Failed to unmarshal request")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errors.ErrJSONParsingFailed)
		return
	}

	if req.State != deviceapi.DeviceEnabled && req.State != deviceapi.DeviceDisabled {
		logger.WithField("device-state", req.State).Error("State provided in request does not match any supported state")
		errMsg := fmt.Sprintf("invalid state. Supported states are %s, %s", deviceapi.DeviceEnabled, deviceapi.DeviceDisabled)
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, errMsg)
		return
	}

	txn, err := transaction.NewTxnWithLocks(ctx, peerID+device)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	err = deviceutils.SetDeviceState(peerID, device, req.State)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Failed to update device state in store")
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, nil)
}

func deviceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	peerID := mux.Vars(r)["peerid"]
	if uuid.Parse(peerID) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "invalid peer-id passed in url")
		return
	}
	deviceName := mux.Vars(r)["device"]
	if deviceName == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Device name not provided in URL")
		return
	}

	// Adding prefix(/) to device name
	deviceName = "/" + deviceName

	txn, err := transaction.NewTxnWithLocks(ctx, peerID+deviceName)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Peer ID not found in store")
		if err == errors.ErrPeerNotFound {
			restutils.SendHTTPError(ctx, w, http.StatusNotFound, errors.ErrPeerNotFound)
		} else {
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "failed to get peer details from store")
		}
		return
	}

	txn.Nodes = []uuid.UUID{peerInfo.ID}
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "delete-device",
			Nodes:  txn.Nodes,
		},
	}

	err = txn.Ctx.Set("peerid", &peerID)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Ctx.Set("device", &deviceName)
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	err = txn.Do()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "transaction to prepare device failed")
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusNoContent, nil)
}

func listAllDevicesHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)
	devices, err := deviceutils.GetDevices()
	if err != nil {
		logger.WithError(err).Error(err)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, devices)
}
