package peercommands

import (
	"net/http"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/rpc/client"

	log "github.com/Sirupsen/logrus"
	etcdclient "github.com/coreos/etcd/client"
	etcdcontext "golang.org/x/net/context"

	"github.com/gorilla/mux"
)

func deletePeerHandler(w http.ResponseWriter, r *http.Request) {
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
	rsp, e := client.ValidateDeletePeer(id, p.Name)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, *rsp.OpError)
		return
	}
	// Delete member from etcd cluster
	c := gdctx.EtcdClient
	mAPI := etcdclient.NewMembersAPI(c)
	e = mAPI.Remove(etcdcontext.Background(), p.MemberID)
	if e != nil {
		log.WithFields(log.Fields{
			"er":   e,
			"peer": id,
		}).Error("Failed to remove member from etcd cluster")

		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	//TODO : Also delete the etcd env from the remote node to be detached?

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
