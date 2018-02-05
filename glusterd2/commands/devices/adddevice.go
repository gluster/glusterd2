package devicecommands

import (
	"context"
	"encoding/json"
	"github.com/pborman/uuid"
	"net/http"

	"github.com/coreos/etcd/clientv3"
	device "github.com/gluster/glusterd2/glusterd2/device"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	log "github.com/sirupsen/logrus"
)

func deviceAddHandler(w http.ResponseWriter, r *http.Request) {
	// Collect inputs from URLi
	ctx := r.Context()
	req := new(device.AddDeviceReq)
	if err := restutils.UnmarshalRequest(r, req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "", api.ErrCodeDefault)
		return
	}
	deviceinfo := device.Info{
		Names:  req.Names,
		PeerID: req.PeerID,
	}
	if req.PeerID == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid Peer ID", api.ErrCodeDefault)
		return
	}

	_, err := store.Store.Get(context.TODO(), deviceinfo.PeerID.String())
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Peer Id not found in store", api.ErrCodeDefault)
		return
	}
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	lock, unlock, err := transaction.CreateLockSteps(deviceinfo.PeerID.String())
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Unable to acquire lock", api.ErrCodeDefault)
		return
	}

	nodes := make([]uuid.UUID, 0)
	nodes = append(nodes, deviceinfo.PeerID)

	txn.Nodes = nodes
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "prepare-device.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("peerid", deviceinfo.PeerID.String())
	txn.Ctx.Set("names", deviceinfo.Names)

	err = txn.Do()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Transaction Failed", api.ErrCodeDefault)
		return
	}
	// update device state
	deviceinfo.State = device.DeviceEnabled
	deviceJSON, err := json.Marshal(deviceinfo)
	if err != nil {
		log.WithField("error", err).Error("Failed to marshal the DeviceInfo object")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to marshal the DeviceInfo object", api.ErrCodeDefault)
		return
	}

	deviceDetails, _ := store.Store.Get(context.TODO(), "devices/"+req.PeerID.String(), clientv3.WithPrefix())

	if len(deviceDetails.Kvs) > 0 {
		for _, kv := range deviceDetails.Kvs {

			var v device.Info

			if err := json.Unmarshal(kv.Value, &v); err != nil {
				restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Unable to add device to store", api.ErrCodeDefault)
				return
			}

			for _, val := range deviceinfo.Names {
				v.Names = append(v.Names, val)
			}
			deviceJSON, err := json.Marshal(v)
			if err != nil {
				restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Unable to add device to store", api.ErrCodeDefault)
				return
			}
			_, err = store.Store.Put(context.TODO(), "devices/"+req.PeerID.String(), string(deviceJSON))
		}
	} else {
		_, err = store.Store.Put(context.TODO(), "devices/"+req.PeerID.String(), string(deviceJSON))
		if err != nil {
			log.WithError(err).Error("Couldn't add deviceinfo to store")
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Unable to add device to store", api.ErrCodeDefault)
			return
		}
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, deviceinfo)

}
