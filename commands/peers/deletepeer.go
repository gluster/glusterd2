package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rest"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func deletePeerHandler(w http.ResponseWriter, r *http.Request) {

	// FIXME: This is not txn based, yet. Behaviour when multiple simultaneous
	// delete peer requests are sent to same node is unknown.

	peerReq := mux.Vars(r)

	id := peerReq["peerid"]
	if id == "" {
		rest.SendHTTPError(w, http.StatusBadRequest, "peerid not present in the request")
		return
	}
	// Check whether the member exists
	p, e := peer.GetPeerF(id)
	if e != nil || p == nil {
		rest.SendHTTPError(w, http.StatusNotFound, "peer not found in cluster")
		return
	}

	// Validate whether the peer can be deleted
	rsp, e := ValidateDeletePeer(id, p.Name)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, rsp.OpError)
		return
	}
	// Delete member from etcd cluster
	e = etcdmgmt.EtcdMemberRemove(p.MemberID)
	if e != nil {
		log.WithFields(log.Fields{
			"er":   e,
			"peer": id,
		}).Error("Failed to remove member from etcd cluster")

		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	// Remove data dir of etcd on remote machine. Restart etcd on remote machine
	// in standalone (single cluster) mode.
	var etcdConf peer.ETCDConfig
	etcdConf.DeletePeer = true
	etcdConf.Name = p.Name
	etcdConf.PeerName = p.Name
	etcdrsp, e := ConfigureRemoteETCD(&etcdConf)
	if e != nil {
		log.WithField("err", e).Error("Failed to configure remote etcd.")
		rest.SendHTTPError(w, http.StatusInternalServerError, etcdrsp.OpError)
		return
	}

	// Remove the peer from the store
	if e := peer.DeletePeer(id); e != nil {
		log.WithFields(log.Fields{
			"er":   e,
			"peer": id,
		}).Error("Failed to remove peer from the store")
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
	} else {
		rest.SendHTTPResponse(w, http.StatusNoContent, nil)
	}

}
