package peer

import (
	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/gdctx"

	log "github.com/Sirupsen/logrus"
)

// AddSelfDetails function adds its own details into the central store
func AddSelfDetails() {
	var memberID uint64

	mlist, e := etcdmgmt.EtcdMemberList()
	if e != nil {
		log.WithField("err", e).Fatal("Failed to list member in etcd cluster")
	}

	for _, memb := range mlist {
		if memb.Name == gdctx.MyUUID.String() {
			memberID = memb.ID
			break
		}
	}
	p := &Peer{
		ID:       gdctx.MyUUID,
		Name:     gdctx.HostName,
		Address:  gdctx.HostIP,
		MemberID: memberID,
	}

	if e = AddOrUpdatePeer(p); e != nil {
		log.WithFields(log.Fields{
			"error":     e,
			"peer/node": p.Name,
		}).Fatal("Failed to add peer into the etcd store")
	}
}
