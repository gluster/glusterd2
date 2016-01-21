package peercommands

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/peer"
	"github.com/gluster/glusterd2/rest"
	"github.com/gluster/glusterd2/rpc/client"
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

	// Add member to etcd server
	var c etcdclient.Client
	mAPI := etcdclient.NewMembersAPI(c)
	//_, e = membersAPI.Add(context.Background(), p.Name)
	member, e := mAPI.Add(etcdcontext.Background(), p.Name)
	if e != nil {
		//rest.SendHTTPError(w, http.StatusInternalServerError, strings.e)
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	newID := member.ID
	newName := member.Name

	log.Info("New member named %s with ID %s added to cluster", newName, newID)

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

	log.Info("ETCD_NAME=%q", newName)
	log.Info("ETCD_INITIAL_CLUSTER=%q", strings.Join(conf, ","))
	log.Info("ETCD_INITIAL_CLUSTER_STATE=\"existing\"")

	if e = peer.AddOrUpdatePeer(p); e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}

	rest.SendHTTPResponse(w, http.StatusOK, nil)

}
