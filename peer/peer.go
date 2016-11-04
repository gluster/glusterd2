// Package peer implements the Peer type
package peer

import (
	"github.com/pborman/uuid"
)

// Peer reperesents a GlusterD
type Peer struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Addresses []string  `json:"addresses"`
	Client    bool      `json:"client"`
	MemberID  uint64    `json:"memberID"`
}

// ETCDConfig represents the structure which holds the ETCD env variables &
// other configurations to be used to set at the remote peer & bring up the etcd
// instance
type ETCDConfig struct {
	PeerName       string
	Name           string
	InitialCluster string
	ClusterState   string
	Client         bool
	DeletePeer     bool
}
