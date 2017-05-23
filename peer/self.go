package peer

import (
	"github.com/gluster/glusterd2/etcdmgmt"
	"github.com/gluster/glusterd2/gdctx"

	config "github.com/spf13/viper"
)

// AddSelfDetails results in the peer adding its own details into etcd
func AddSelfDetails() error {

	mlist, err := etcdmgmt.EtcdMemberList()
	if err != nil {
		return err
	}

	var memberID uint64
	for _, memb := range mlist {
		if memb.Name == gdctx.MyUUID.String() {
			memberID = memb.ID
			break
		}
	}

	p := &Peer{
		ID:        gdctx.MyUUID,
		Name:      gdctx.HostName,
		Addresses: []string{config.GetString("peeraddress")},
		MemberID:  memberID,
	}

	return AddOrUpdatePeer(p)
}
