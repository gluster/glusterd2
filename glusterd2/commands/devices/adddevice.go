package devicecommands

import (
	"net/http"
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/device"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
	"github.com/gorilla/mux"
)

func deviceAddHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	req := new(api.AddDeviceReq)
	if err := restutils.UnmarshalRequest(r, req); err != nil {
		logger.WithError(err).Error("Failed to Unmarshal request")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Unable to marshal request", api.ErrCodeDefault)
		return
	}
	peerID := mux.Vars(r)["peerid"]
	if peerID == "" {
                restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "peerid not present in request", api.ErrCodeDefault)
                return
        }
	fmt.Printf("Printing peer ID %s", peerID)
	p, err := peer.GetPeer(peerID)
	if err != nil {
                logger.WithError(err).WithField("peerid", peerID).Error("Peer ID not found in store")
                restutils.SendHTTPError(ctx, w, http.StatusNotFound, "Peer Id not found in store", api.ErrCodeDefault)
                return
        }
	var v []api.DeviceInfo
	fmt.Printf("Printing Peer  %s",p)
	for _, name := range req.Devices {
		tempInfo := api.DeviceInfo{
			Name: name,
		}
		v = append(v, tempInfo)
	}
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()

	nodes := make([]uuid.UUID, 0)
	nodes = append(nodes, uuid.UUID(peerID))

	txn.Nodes = nodes
	txn.Steps = []*transaction.Step{
		{
			DoFunc: "prepare-device.Commit",
			Nodes:  txn.Nodes,
		},
	}
	txn.Ctx.Set("peerid", peerID)
	txn.Ctx.Set("device-details", v)

	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("Transaction Failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Transaction Failed", api.ErrCodeDefault)
		return
	}
	deviceInfo, _ := device.GetDevice(peerID)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, deviceInfo)
}
