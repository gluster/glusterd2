package devicecommands

import (
	"context"
	"encoding/json"
	"net/http"
	"github.com/pborman/uuid"

        device "github.com/gluster/glusterd2/glusterd2/device"
        "github.com/coreos/etcd/clientv3"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/store"
        "github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/glusterd2/transaction"
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
	nodeIDRaw := req.NodeID
	deviceName := req.DeviceName
	var deviceinfo device.DeviceInfo
        deviceinfo.DeviceName = deviceName
	nodeID := nodeIDRaw
	if nodeID == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid Node ID", api.ErrCodeDefault)
		return
	}
	deviceinfo.NodeID = nodeID
	_, err := store.Store.Get(context.TODO(), deviceinfo.NodeID.String())
	if err != nil {
                
		restutils.SendHTTPError(ctx,w, http.StatusInternalServerError, "Node Id not found in store", api.ErrCodeDefault)
		return
	}
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	lock, unlock, err := transaction.CreateLockSteps(deviceinfo.NodeID.String())
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Unable to acquire lock", api.ErrCodeDefault)
		return
	}

	nodes := make([]uuid.UUID, 0)
	nodes = append(nodes, deviceinfo.NodeID)

	txn.Nodes = nodes
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "prepare-device.Commit",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	txn.Ctx.Set("nodeid", deviceinfo.NodeID.String())
	txn.Ctx.Set("devicename", deviceinfo.DeviceName)

	err = txn.Do()
	if err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Transaction Failed", api.ErrCodeDefault)
		return
	}
	// update device state
	deviceinfo.State = device.DeviceEnabled
	json1, err := json.Marshal(deviceinfo)
	if err != nil {
		log.WithField("error", err).Error("Failed to marshal the DeviceInfo object")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to marshal the DeviceInfo object", api.ErrCodeDefault)
		return
	}

       deviceDetails, _ := store.Store.Get(context.TODO(), "devices/" + nodeID.String(), clientv3.WithPrefix())

       if len(deviceDetails.Kvs) > 0 {
           for _, kv := range deviceDetails.Kvs {
               
               var v device.DeviceInfo
               
               _ = json.Unmarshal(kv.Value, &v)

               for _, val := range deviceinfo.DeviceName {
                   v.DeviceName = append(v.DeviceName, val)                    
               }
               json2, _ := json.Marshal(v)
               _, err = store.Store.Put(context.TODO(), "devices/" + nodeID.String(), string(json2))
           }
       } else {
	    _, err = store.Store.Put(context.TODO(), "devices/" + nodeID.String(), string(json1))
	    if err != nil {
		log.WithError(err).Error("Couldn't add deviceinfo to store")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Unable to add device to store", api.ErrCodeDefault)
		return
	    }
        }
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, deviceinfo)
        
}

