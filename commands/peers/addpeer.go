package peercommands

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/rpc/client"
	"github.com/gluster/glusterd2/rpc/services"
	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	etcdclient "github.com/coreos/etcd/client"
	etcdcontext "golang.org/x/net/context"

	"github.com/pborman/uuid"
)

func addPeerHandler(w http.ResponseWriter, r *http.Request) {
	var req peer.PeerAddRequest
	if e := utils.GetJSONFromRequest(r, &req); e != nil {
		rest.SendHTTPError(w, http.StatusBadRequest, e.Error())
		return
	}

	if len(req.Addresses) < 1 {
		rest.SendHTTPError(w, http.StatusBadRequest, errors.ErrNoHostnamesPresent.Error())
		return
	}

	if req.Name == "" {
		req.Name = req.Addresses[0]
	}

	//TODO: Do proper validation before initiating the add process

	//FIXME: In the correct add process, the peer being probed would add it's details to the store once it's been validated. The code below is just a temporary stand-in to show how the API's would work

	p := &peer.Peer{
		ID:        uuid.NewRandom(),
		Name:      req.Name,
		Addresses: req.Addresses,
	}

	rsp, e := client.ValidateAddPeer(&req)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, *rsp.OpError)
		return
	}

	var c etcdclient.Client
	// Add member to etcd server
	mAPI := etcdclient.NewMembersAPI(c)
	member, e := mAPI.Add(etcdcontext.Background(), p.Name)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	newID := member.ID
	newName := member.Name

	log.WithFields(log.Fields{
		"New member ": newName,
		"member Id ":  newID,
	}).Info("New member added to the cluster")

	mlist, e := mAPI.List(etcdcontext.Background())
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	conf := []string{}
	for _, memb := range mlist {
		for _, u := range memb.PeerURLs {
			n := memb.Name
			if memb.ID == newID {
				n = newName
			}
			conf = append(conf, fmt.Sprintf("%s=%s", n, u))
		}
	}

	log.WithField("ETCD_NAME=", newName).Info("ETCD_NAME")
	log.WithField("ETCD_INITIAL_CLUSTER=", strings.Join(conf, ",")).Info("ETCD_INITIAL_CLUSTER")
	log.Info("ETCD_INITIAL_CLUSTER_STATE=\"existing\"")

	var etcdenv services.RPCEtcdEnvReq
	*etcdenv.PeerName = p.Name
	*etcdenv.Name = newName
	*etcdenv.InitialCluster = strings.Join(conf, ",")
	*etcdenv.ClusterState = "existing"

	etcdrsp, e := client.AddEtcdEnvVar(&etcdenv)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, *etcdrsp.OpError)
		return
	}

	if e = peer.AddOrUpdatePeer(p); e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	rest.SendHTTPResponse(w, http.StatusOK, nil)

}
