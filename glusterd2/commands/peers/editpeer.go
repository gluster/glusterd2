package peercommands

import (
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func editPeer(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	var req api.PeerEditReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, err.Error(), api.ErrCodeDefault)
		return
	}
	for k := range req.MetaData {
		if strings.HasPrefix(k, "_") {
			logger.WithField("metadata-key", req.MetaData).Error("Key names starting with '_' are restricted in Metadata field")
			restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Key names starting with '_' are restricted in Metadata field", api.ErrCodeDefault)
			return
		}
	}
	peerID := mux.Vars(r)["peerid"]
	if uuid.Parse(peerID) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid peerID passed in url", api.ErrCodeDefault)
		return
	}

	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(string(peerID))
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Failed to get lock on peer")
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, "Failed to get lock on peer", api.ErrCodeDefault)
		return
	}

	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "peer-edit",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
		unlock,
	}
	err = txn.Ctx.Set("peerid", peerID)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Failed to set peerID in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	err = txn.Ctx.Set("req", req)
	if err != nil {
		logger.WithError(err).WithField("req-key", req).Error("Failed to set unmarshalled request information in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("Transaction to update peer failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Transaction to update metadata failed", api.ErrCodeDefault)
		return
	}
	var peerInfo peer.Peer
	if err := txn.Ctx.Get("peerInfo", &peerInfo); err != nil {
		logger.WithError(err).Error("Failed to retrieve peer information in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to retrieve peer information in transaction context", api.ErrCodeDefault)
		return
	}
	resp := createPeerEditResp(&peerInfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusCreated, resp)

}

func txnPeerEdit(c transaction.TxnCtx) error {
	var peerID string
	if err := c.Get("peerid", &peerID); err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Failed transaction, cannot fetch peer-id from transaction context")
		return err
	}

	var req api.PeerEditReq
	if err := c.Get("req", &req); err != nil {
		c.Logger().WithError(err).WithField("req-key", req).Error("Failed transaction, cannot fetch request-information from transaction context")
		return err
	}
	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Peer ID not found in store")
		return err
	}
	for k, v := range req.MetaData {
		if peerInfo.MetaData != nil {
			peerInfo.MetaData[k] = v
		} else {
			peerInfo.MetaData = make(map[string]string)
			peerInfo.MetaData[k] = v
		}
	}
	err = peer.AddOrUpdatePeer(peerInfo)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Failed to update peer Info")
		return err
	}
	err = c.Set("peerInfo", peerInfo)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Failed to set peer info in transaction context")
		return err
	}
	return nil
}

func registerPeerEditStepFuncs() {
	var sfs = []struct {
		name string
		sf   transaction.StepFunc
	}{
		{"peer-edit", txnPeerEdit},
	}
	for _, sf := range sfs {
		transaction.RegisterStepFunc(sf.sf, sf.name)
	}
}

func createPeerEditResp(p *peer.Peer) *api.PeerEditResp {
	return &api.PeerEditResp{
		ID:              p.ID,
		Name:            p.Name,
		PeerAddresses:   p.PeerAddresses,
		ClientAddresses: p.ClientAddresses,
		MetaData:        p.MetaData,
	}
}
