package peercommands

import (
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/store"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
)

func addPeerHandler(w http.ResponseWriter, r *http.Request) {
	var req PeerAddReq
	if e := utils.GetJSONFromRequest(r, &req); e != nil {
		restutils.SendHTTPError(w, http.StatusBadRequest, e.Error())
		return
	}

	if len(req.Addresses) < 1 {
		restutils.SendHTTPError(w, http.StatusBadRequest, errors.ErrNoHostnamesPresent.Error())
		return
	}
	log.WithField("addresses", req.Addresses).Debug("recieved request to add new peer with given addresses")

	p, _ := peer.GetPeerByAddrs(req.Addresses)
	if p != nil {
		restutils.SendHTTPError(w, http.StatusConflict, fmt.Sprintf("Peer exists with given addresses (ID: %s)", p.ID.String()))
		return
	}

	// A peer can have multiple addresses. For now, we use only the first
	// address present in the req.Addresses list.

	remotePeerAddress, err := utils.FormRemotePeerAddress(req.Addresses[0])
	if err != nil {
		log.WithError(err).WithField("remote", remotePeerAddress).Error("failed to grpc.Dial remote")
		restutils.SendHTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	client, err := GetPeerServiceClient(remotePeerAddress)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer client.conn.Close()

	// This remote call will return the remote peer's ID (UUID), name
	remotePeer, err := client.ValidateAddPeer(&req)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	} else if remotePeer.OpRet != 0 {
		restutils.SendHTTPError(w, http.StatusInternalServerError, remotePeer.OpError)
		return
	}

	newconfig := &StoreConfig{store.Store.Endpoints()}
	log.WithFields(log.Fields{
		"peer":      remotePeer.UUID,
		"endpoints": newconfig.Endpoints,
	}).Debug("asking new peer to join cluster with given endpoints")
	rsp, err := client.JoinCluster(newconfig)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	} else if rsp.OpRet != 0 {
		restutils.SendHTTPError(w, http.StatusInternalServerError, rsp.OpError)
		return
	}

	newpeer, err := peer.GetPeer(remotePeer.UUID)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, "new peer was added, but could not find peer in store. Try againg later.")
	} else {
		restutils.SendHTTPResponse(w, http.StatusCreated, newpeer)
	}
}
