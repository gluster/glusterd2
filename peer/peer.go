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
	MemberID  string    `json:"memberID"`
}

// PeerAddRequest represents the structure to be added into the store
type PeerAddRequest struct {
	Addresses []string `json:"addresses"`
	Name      string   `json:"name,omitempty"`
	Client    bool     `json:"client,omitempty"`
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
}
