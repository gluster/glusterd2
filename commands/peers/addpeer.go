package peercommands

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
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

	local_node := false
	for _, addr := range req.Addresses {
		local, _ := utils.IsLocalAddress(addr)
		if local == true {
			local_node = true
			break
		}
	}

	if local_node == true {
		rest.SendHTTPError(w, http.StatusInternalServerError, errors.ErrPeerLocalNode.Error())
		return
	}

	rsp, e := client.ValidateAddPeer(&req)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, *rsp.OpError)
		return
	}

	var etcdConf peer.ETCDConfig
	p := &peer.Peer{
		ID:        uuid.Parse(*rsp.UUID),
		Name:      req.Name,
		Addresses: req.Addresses,
		MemberID:  "",
	}

	if req.Client == false {
		c := gdctx.EtcdClient
		// Add member to etcd server
		mAPI := etcdclient.NewMembersAPI(c)
		member, e := mAPI.Add(etcdcontext.Background(), "http://"+req.Name+":2380")
		if e != nil {
			log.WithFields(log.Fields{
				"error":  e,
				"member": req.Name,
			}).Error("Failed to add member into etcd cluster")
			rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
			return
		}

		p.MemberID = member.ID
		newName := "ETCD_" + p.Name

		log.WithFields(log.Fields{
			"New member ": newName,
			"member Id ":  member.ID,
		}).Info("New member added to the cluster")

		mlist, e := mAPI.List(etcdcontext.Background())
		if e != nil {
			log.WithField("err", e).Error("Failed to list member in etcd cluster")
			rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
			return
		}

		conf := []string{}
		for _, memb := range mlist {
			for _, u := range memb.PeerURLs {
				n := memb.Name
				if memb.ID == p.MemberID {
					n = newName
				}
				conf = append(conf, fmt.Sprintf("%s=%s", n, u))
			}
		}

		log.WithField("ETCD_NAME", newName).Info("ETCD_NAME")
		log.WithField("ETCD_INITIAL_CLUSTER", strings.Join(conf, ",")).Info("ETCD_INITIAL_CLUSTER")
		log.Info("ETCD_INITIAL_CLUSTER_STATE\"existing\"")

		etcdConf.Name = newName
		etcdConf.InitialCluster = strings.Join(conf, ",")
		etcdConf.ClusterState = "existing"
	} else {
		etcdConf.Name = "ETCD_" + req.Name
		initialCluster, err := peer.GetInitialCluster()
		if err != nil {
			log.WithField("err", e).Error("Failed to construct initialCluster")
			rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		}
		etcdConf.InitialCluster = initialCluster
		etcdConf.ClusterState = ""
	}

	etcdConf.Client = req.Client
	etcdConf.PeerName = req.Name
	etcdrsp, e := client.ConfigureRemoteETCD(&etcdConf)
	if e != nil {
		log.WithField("err", e).Error("Failed to configure remote etcd")
		rest.SendHTTPError(w, http.StatusInternalServerError, *etcdrsp.OpError)
		return
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
