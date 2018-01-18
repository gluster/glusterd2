package heketi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pborman/uuid"

	heketiapi "github.com/gluster/glusterd2/plugins/heketi/api"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
        "github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/glusterd2/transaction"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	heketiPrefix string = store.GlusterPrefix + "heketi/"
)

func heketiDeviceAddHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URL
	p := mux.Vars(r)
	nodeIDRaw := p["nodeid"]
	deviceName := p["devicename"]
	var deviceinfo heketiapi.DeviceInfo
	deviceinfo.DeviceName = deviceName
        ctx := r.Context()
	nodeID := uuid.Parse(nodeIDRaw)
	if nodeID == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid Node ID", api.ErrCodeDefault)
		return
	}
	deviceinfo.NodeID = nodeID

	reqID, logger := restutils.GetReqIDandLogger(r)

	_, err := store.Store.Get(context.TODO(), deviceinfo.NodeID.String())
	if err != nil {
		restutils.SendHTTPError(ctx,w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	lock, unlock, err := transaction.CreateLockSteps(deviceinfo.NodeID.String())
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	nodes := make([]uuid.UUID, 0)
	nodes = append(nodes, deviceinfo.NodeID)

	txn.Nodes = nodes
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "heketi-prepare-device.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("nodeid", deviceinfo.NodeID.String())
	txn.Ctx.Set("devicename", deviceinfo.DeviceName)

	err = txn.Do()
	if err != nil {
		logger.WithFields(log.Fields{
			"error":      err.Error(),
			"nodeid":     deviceinfo.NodeID,
			"devicename": deviceinfo.DeviceName,
		}).Error("Failed to prepare device")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	// update device state
	deviceinfo.State = heketiapi.HeketiDeviceEnabled

	json, err := json.Marshal(deviceinfo)
	if err != nil {
		log.WithField("error", err).Error("Failed to marshal the DeviceInfo object" + reqID)
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	_, err = store.Store.Put(context.TODO(), heketiPrefix+"/"+deviceinfo.NodeID.String()+"/"+deviceName, string(json))
	if err != nil {
		log.WithError(err).Error("Couldn't add deviceinfo to store")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}

	restutils.SendHTTPResponse(ctx, w, http.StatusOK, deviceinfo)
}
