package devicecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/device"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
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
	deviceinfo := api.Device{
		PeerID: req.PeerID,
	}
	for _, name := range req.Names {
		tempInfo := api.Info{
			Name: name,
		}
		deviceinfo.Detail = append(deviceinfo.Detail, tempInfo)
	}
	if req.PeerID == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Peer ID not found in request", api.ErrCodeDefault)
		return
	}

	_, err := peer.GetPeer(deviceinfo.PeerID.String())
	if err != nil {
		logger.WithError(err).WithField("peerid", req.PeerID).Error("Peer ID not found in store")
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, "Peer Id not found in store", api.ErrCodeDefault)
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
	txn.Ctx.Set("device-details", deviceinfo.Detail)

	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("Transaction Failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Transaction Failed", api.ErrCodeDefault)
		return
	}
	deviceInfo, _ := device.GetDevice(req.PeerID.String())
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, deviceInfo)
}
