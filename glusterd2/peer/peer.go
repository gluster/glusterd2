// Package peer implements the Peer type
package peer

import (
	"strings"

	"github.com/pborman/uuid"
)

// Peer reperesents a GlusterD
type Peer struct {
	ID              uuid.UUID
	Name            string
	PeerAddresses   []string
	ClientAddresses []string
	Metadata        map[string]string
}

// ETCDConfig represents the structure which holds the ETCD env variables &
// other configurations to be used to set at the remote peer & bring up the etcd
// instance
type ETCDConfig struct {
	PeerName       string
	Name           string
	InitialCluster string
	ClusterState   string
	DeletePeer     bool
}

// MetadataSize returns the size of metadata from peer info
func (p *Peer) MetadataSize() int {
	size := 0
	for key, value := range p.Metadata {
		if !strings.HasPrefix(key, "_") {
			size = size + len(key) + len(value)
		}
	}
	return size
}
