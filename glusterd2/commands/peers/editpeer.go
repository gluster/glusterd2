package peercommands

import (
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"
	gderrors "github.com/gluster/glusterd2/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

func editPeer(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	var req api.PeerEditReq
	if err := restutils.UnmarshalRequest(r, &req); err != nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrJSONParsingFailed)
		return
	}

	peerID := mux.Vars(r)["peerid"]
	if uuid.Parse(peerID) == nil {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Invalid peerID passed in url")
		return
	}

	for key := range req.Metadata {
		if strings.HasPrefix(key, "_") {
			logger.WithError(gderrors.ErrRestrictedKeyFound).WithField("metadata-key", key)
			restutils.SendHTTPError(ctx, w, http.StatusBadRequest, gderrors.ErrRestrictedKeyFound)
			return
		}
	}

	txn, err := transaction.NewTxnWithLocks(ctx, peerID)
	if err != nil {
		status, err := restutils.ErrToStatusCode(err)
		restutils.SendHTTPError(ctx, w, status, err)
		return
	}
	defer txn.Done()

	txn.Steps = []*transaction.Step{
		{
			DoFunc: "peer-edit",
			Nodes:  []uuid.UUID{gdctx.MyUUID},
		},
	}
	err = txn.Ctx.Set("peerid", peerID)
	if err != nil {
		logger.WithError(err).WithFields(
			log.Fields{"key": "peerid", "value": peerID}).Error("Failed to set key in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	err = txn.Ctx.Set("req", req)
	if err != nil {
		logger.WithError(err).WithField("key", "req").Error("Failed to set key in transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
		return
	}
	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("Transaction to update peer failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Transaction to update peer failed")
		return
	}
	var peerInfo peer.Peer
	if err := txn.Ctx.Get("peerInfo", &peerInfo); err != nil {
		logger.WithError(err).WithField("key", "peerInfo").Error("Failed to get key from transaction context")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Failed to get key from  transaction context")
		return
	}
	resp := createPeerEditResp(&peerInfo)
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, resp)

}

func txnPeerEdit(c transaction.TxnCtx) error {
	var peerID string
	if err := c.Get("peerid", &peerID); err != nil {
		c.Logger().WithError(err).WithField("key", "peerID").Error("Failed to get key from transaction context")
		return err
	}

	var req api.PeerEditReq
	if err := c.Get("req", &req); err != nil {
		c.Logger().WithError(err).WithField("key", "req").Error("Failed to get key from transaction context")
		return err
	}
	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Peer ID not found in store")
		return err
	}

	if req.MetadataSize() > maxMetadataSizeLimit {
		return gderrors.ErrMetadataSizeOutOfBounds
	}

	for k, v := range req.Metadata {
		peerInfo.Metadata[k] = v
	}

	if peerInfo.MetadataSize() > maxMetadataSizeLimit {
		return gderrors.ErrMetadataSizeOutOfBounds
	}

	if req.Zone != "" {
		peerInfo.Metadata["_zone"] = req.Zone
	}
	err = peer.AddOrUpdatePeer(peerInfo)
	if err != nil {
		c.Logger().WithError(err).WithField("peerid", peerID).Error("Failed to update peer Info")
		return err
	}
	err = c.Set("peerInfo", peerInfo)
	if err != nil {
		c.Logger().WithError(err).WithField("key", "peerInfo").Error("Failed to set key in transaction context")
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
		Metadata:        p.Metadata,
	}
}
