package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func deletePeerHandler(w http.ResponseWriter, r *http.Request) {

	// FIXME: This is not txn based, yet. Behaviour when multiple simultaneous
	// delete peer requests are sent to same node is unknown.

	peerReq := mux.Vars(r)

	id := peerReq["peerid"]
	if id == "" {
		restutils.SendHTTPError(w, http.StatusBadRequest, "peerid not present in the request")
		return
	}
	// Check whether the member exists
	p, e := peer.GetPeerF(id)
	if e != nil || p == nil {
		restutils.SendHTTPError(w, http.StatusNotFound, "peer not found in cluster")
		return
	}

	// Removing self should be disallowed (like in glusterd1)
	if id == gdctx.MyUUID.String() {
		restutils.SendHTTPError(w, http.StatusBadRequest, "Removing self is disallowed.")
		return
	}

	remotePeerAddress, err := utils.FormRemotePeerAddress(p.Addresses[0])
	if err != nil {
		restutils.SendHTTPError(w, http.StatusBadRequest, err.Error())
		return
	}

	client, err := getPeerServiceClient(remotePeerAddress)
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer client.conn.Close()

	// Validate whether the peer can be deleted
	rsp, e := client.ValidateDeletePeer(id)
	if e != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, rsp.OpError)
		return
	}

	// Remove the peer from the store
	if e := peer.DeletePeer(id); e != nil {
		log.WithFields(log.Fields{
			"er":   e,
			"peer": id,
		}).Error("Failed to remove peer from the store")
		restutils.SendHTTPError(w, http.StatusInternalServerError, e.Error())
	} else {
		restutils.SendHTTPResponse(w, http.StatusNoContent, nil)
	}

	rsp, err = client.JoinCluster(&StoreConfig{})
	if err != nil {
		restutils.SendHTTPError(w, http.StatusInternalServerError, err.Error())
	} else if rsp.OpRet != 0 {
		restutils.SendHTTPError(w, http.StatusInternalServerError, rsp.OpError)
	}
}
