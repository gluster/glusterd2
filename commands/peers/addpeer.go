package peercommands

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/context"
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
		Client:    req.Client,
	}

	rsp, e := client.ValidateAddPeer(&req)
	if e != nil {
		rest.SendHTTPError(w, http.StatusInternalServerError, *rsp.OpError)
		return
	}

	var etcdConf peer.ETCDConfig
	log.Info("In peer add")
	if req.Client == false {
		c := context.EtcdClient
		// Add member to etcd server
		mAPI := etcdclient.NewMembersAPI(c)
		member, e := mAPI.Add(etcdcontext.Background(), "http://"+p.Name+":2380")
		if e != nil {
			log.WithFields(log.Fields{
				"error":  e,
				"member": p.Name,
			}).Error("Failed to add member into etcd cluster")
			rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
			return
		}

		newID := member.ID
		newName := "ETCD_" + p.Name

		log.WithFields(log.Fields{
			"New member ": newName,
			"member Id ":  newID,
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
				if memb.ID == newID {
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
		log.Debug("Calling GetInitialCluster")
		etcdConf.PeerName = ""
		etcdConf.Name = ""
		initialCluster, err := peer.GetInitialCluster()
		if err != nil {
			log.WithField("err", e).Error("Failed to construct initialCluster")
			rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		}
		log.Debug("InitialCluster", initialCluster)
		etcdConf.InitialCluster = initialCluster
		etcdConf.ClusterState = ""
	}
	etcdConf.Client = req.Client
	etcdConf.PeerName = p.Name
	log.Debug("Calling client.ConfigureRemoteETCD")
	etcdrsp, e := client.ConfigureRemoteETCD(&etcdConf)
	if e != nil {
		log.WithField("err", e).Error("Failed to configure remote etcd")
		rest.SendHTTPError(w, http.StatusInternalServerError, *etcdrsp.OpError)
		return
	}
	log.Debug("client.ConfigureRemoteETCD is called")
	if e = peer.AddOrUpdatePeer(p); e != nil {
		log.WithFields(log.Fields{
			"error":     e,
			"peer/node": p.Name,
		}).Error("Failed to add peer into the etcd store")
		rest.SendHTTPError(w, http.StatusInternalServerError, e.Error())
		return
	}
	log.Debug("peer.AddOrUpdatePeer is called")
	rest.SendHTTPResponse(w, http.StatusOK, nil)
}
