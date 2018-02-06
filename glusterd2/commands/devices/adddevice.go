package devicecommands

import (
	"context"
	"encoding/json"
	"net/http"

	device "github.com/gluster/glusterd2/glusterd2/device"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"

	log "github.com/sirupsen/logrus"
	"github.com/pborman/uuid"
	"github.com/coreos/etcd/clientv3"
)

func deviceAddHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	req := new(api.AddDeviceReq)
	if err := restutils.UnmarshalRequest(r, req); err != nil {
		logger.WithError(err).WithField("devices", "devices").Error("Failed to marshal request")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Unable to marshal request", api.ErrCodeDefault)
		return
	}
	deviceinfo := api.Info{
		Names:  req.Names,
		PeerID: req.PeerID,
	}
	if req.PeerID == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Peer ID not found in request", api.ErrCodeDefault)
		return
	}

	_, err := peer.GetPeer(deviceinfo.PeerID.String())
	if err != nil {
		logger.WithError(err).WithField("peerid", req.PeerID).Error("Peer ID not found in store")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Peer Id not found in store", api.ErrCodeDefault)
		return
	}
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	nodes := make([]uuid.UUID, 0)
	nodes = append(nodes, deviceinfo.PeerID)

	txn.Nodes = nodes
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "prepare-device.Commit",
			Nodes:  txn.Nodes,
		},
	}
	txn.Ctx.Set("peerid", deviceinfo.PeerID.String())
	txn.Ctx.Set("names", deviceinfo.Names)

	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("Transaction Failed")
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

			var v api.Info

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
