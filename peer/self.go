package peer

import (
	"github.com/gluster/glusterd2/gdctx"

	log "github.com/Sirupsen/logrus"
	etcdclient "github.com/coreos/etcd/client"
	etcdcontext "golang.org/x/net/context"
)

// AddSelfDetails function adds its own details into the central store
func AddSelfDetails() {
	var memberID string
	c := gdctx.EtcdClient
	mAPI := etcdclient.NewMembersAPI(c)

	mlist, e := mAPI.List(etcdcontext.Background())
	if e != nil {
		log.WithField("err", e).Fatal("Failed to list member in etcd cluster")
	}

	for _, memb := range mlist {
		for _, _ = range memb.PeerURLs {
			if memb.Name == "default" {
				memberID = memb.ID
				break
			}
		}
	}
	p := &Peer{
		ID:        gdctx.MyUUID,
		Name:      gdctx.HostIP,
		Addresses: []string{gdctx.HostIP},
		MemberID:  memberID,
	}

	if e = AddOrUpdatePeer(p); e != nil {
		log.WithFields(log.Fields{
			"error":     e,
			"peer/node": p.Name,
		}).Fatal("Failed to add peer into the etcd store")
	}
}
