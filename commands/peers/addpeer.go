package peercommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/api"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/store"
	"github.com/gluster/glusterd2/utils"

	log "github.com/sirupsen/logrus"
)

type peerAddReq struct {
	Addresses []string
}

func addPeerHandler(w http.ResponseWriter, r *http.Request) {
	var req peerAddReq
	if e := utils.GetJSONFromRequest(r, &req); e != nil {
		restutils.SendHTTPError(w, http.StatusBadRequest, e.Error(), api.ErrCodeDefault)
		return
	}

	if len(req.Addresses) < 1 {
		restutils.SendHTTPError(w, http.StatusBadRequest, errors.ErrNoHostnamesPresent.Error(), api.ErrCodeDefault)
		return
	}
	log.WithField("addresses", req.Addresses).Debug("received request to add new peer with given addresses")

	p, _ := peer.GetPeerByAddrs(req.Addresses)
	if p != nil {
		restutils.SendHTTPError(w, http.StatusConflict, fmt.Sprintf("Peer exists with given addresses (ID: %s)", p.ID.String()), api.ErrCodeDefault)
		return
	}

	// A peer can have multiple addresses. For now, we use only the first
	// address present in the req.Addresses list.
	remotePeerAddress, err := utils.FormRemotePeerAddress(req.Addresses[0])
	if err != nil {
		log.WithError(err).WithField("address", req.Addresses[0]).Error("failed to parse peer address")
		restutils.SendHTTPError(w, http.StatusBadRequest, "failed to parse remote address", api.ErrCodeDefault)
		return
	}

	// TODO: Try all addresses till the first one connects
	client, err := getPeerServiceClient(remotePeerAddress)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	defer client.conn.Close()
	logger := log.WithField("peer", remotePeerAddress)

	newconfig := &StoreConfig{store.Store.Endpoints()}
	logger.WithField("endpoints", newconfig.Endpoints).Debug("asking new peer to join cluster with given endpoints")

	// Ask the peer to join the cluster
	rsp, err := client.JoinCluster(newconfig)
	if err != nil {
		log.WithError(err).Error("sending Join request failed")
		restutils.SendHTTPError(w, http.StatusInternalServerError, "failed to send join cluster request", api.ErrCodeDefault)
		return
	} else if Error(rsp.Err) != ErrNone {
		err = Error(rsp.Err)
		logger.WithError(err).Error("join request failed")
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	logger = logger.WithField("peerid", rsp.PeerID)
	logger.Info("new peer joined our cluster")

	// Get the new peer information to reply back with
	newpeer, err := peer.GetPeer(rsp.PeerID)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "new peer was added, but could not find peer in store. Try again later.", api.ErrCodeDefault)
		return
	}

	resp := createPeerAddResp(newpeer)
	restutils.SendHTTPResponse(w, http.StatusCreated, resp)

	// Save updated store endpoints for restarts
	store.Store.UpdateEndpoints()
}

func createPeerAddResp(p *peer.Peer) *api.PeerAddResp {
	return &api.PeerAddResp{
		ID:        p.ID,
		Name:      p.Name,
		Addresses: p.Addresses,
	}
}
