package peercommands

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

func addPeerHandler(w http.ResponseWriter, r *http.Request) {

	// FIXME: This is not txn based, yet. Behaviour when multiple simultaneous
	// add peer requests are sent to same node is unknown.

	var req PeerAddReq
	if e := utils.GetJSONFromRequest(r, &req); e != nil {
		rest.SendHTTPError(w, http.StatusBadRequest, e.Error())
		return
	}

	if len(req.Addresses) < 1 {
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrNoHostnamesPresent.Error())
		return
	}

	localNode := false
	for _, addr := range req.Addresses {
		local, _ := utils.IsLocalAddress(addr)
		if local == true {
			localNode = true
			break
		}
	}

	if localNode == true {
		rest.SendHTTPError(w, http.StatusInternalServerError, errors.ErrPeerLocalNode.Error())
		return
	}

	// This remote call will return the remote peer's ID (UUID) and Name.
	remotePeer, e := ValidateAddPeer(&req)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, remotePeer.OpError)
		return
	}

	// If user hasn't specified peer name, use the name returned by remote
	// peer which defaults to it's hostname.
	if req.Name == "" {
		req.Name = remotePeer.PeerName
	}

	var etcdConf EtcdConfigReq
	etcdConf.EtcdName = remotePeer.UUID
	etcdConf.Client = req.Client

	var newMemberID uint64

	// By default, req.Client is false. This means every new node added via
	// add peer will be a member in etcd cluster and participate in
	// consensus. TODO: This name "client" in the REST API should really be
	// changed! May be to etcdproxy or just proxy ?
	if req.Client == false {

		// Adding a member is a two step process:
		// 	1. Add the new member to the cluster via the members API. This is
		//	   performed on this node i.e the one that just accepted peer add
		//	   request from the user.
		//	2. Start the new member on the target node (the new peer) with the new
		//         cluster configuration, including a list of the updated members
		//	   (existing members + the new member).

		newMember, e := etcdmgmt.EtcdMemberAdd("http://" + req.Addresses[0] + ":2380")
		if e != nil {
			log.WithFields(log.Fields{
				"error":   e,
				"uuid":    remotePeer.UUID,
				"name":    req.Name,
				"address": req.Addresses[0],
			}).Error("Failed to add member to etcd cluster.")
			rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
			return
		}
		newMemberID = newMember.ID

		log.WithField("member-id", newMemberID).Info("Added new member to etcd cluster")

		mlist, e := etcdmgmt.EtcdMemberList()
		if e != nil {
			log.WithField("err", e).Error("Failed to list members in etcd cluster")
			rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
			return
		}

		// Member name of the newly added etcd member has not been set at this point.
		conf := []string{}
		for _, memb := range mlist {
			for _, u := range memb.PeerURLs {
				n := memb.Name
				if memb.ID == newMemberID {
					n = remotePeer.UUID
				}
				conf = append(conf, fmt.Sprintf("%s=%s", n, u))
			}
		}

		etcdConf.InitialCluster = strings.Join(conf, ",")
		etcdConf.ClusterState = "existing"
	} else {
		// Run etcd on remote node in proxy mode. embed does not support this yet.
	}

	log.WithField("initial-cluster", etcdConf.InitialCluster).Debug("Reconfiguring etcd on remote peer")

	etcdrsp, e := ConfigureRemoteETCD(req.Addresses[0], &etcdConf)
	if e != nil {
		log.WithField("err", e).Error("Failed to configure remote etcd")
		rest.SendHTTPError(w, http.StatusInternalServerError, etcdrsp.OpError)
		return
	}

	// Create a new peer object and add it the store,
	p := &peer.Peer{
		ID:        uuid.Parse(remotePeer.UUID),
		Name:      req.Name,
		Addresses: req.Addresses,
		MemberID:  newMemberID,
	}
	if e = peer.AddOrUpdatePeer(p); e != nil {
		log.WithFields(log.Fields{
			"error":     e,
			"peer/node": p.Name,
		}).Error("Failed to add peer into the etcd store")
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	body := map[string]uuid.UUID{"id": p.ID}
	rest.SendHTTPResponse(w, http.StatusCreated, body)
}
