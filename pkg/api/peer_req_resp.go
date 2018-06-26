package api

import (
	"github.com/pborman/uuid"
)

// Peer represents a peer in the glusterd2 cluster
type Peer struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	PeerAddresses   []string          `json:"peer-addresses"`
	ClientAddresses []string          `json:"client-addresses"`
	Online          bool              `json:"online"`
	PID             int               `json:"pid,omitempty"`
	Metadata        map[string]string `json:"metadata"`
}

// PeerAddReq represents an incoming request to add a peer to the cluster
type PeerAddReq struct {
	Addresses []string          `json:"addresses"`
	Zone      string            `json:"zone,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// PeerEditReq represents an incoming request to edit metadata of peer
type PeerEditReq struct {
	Zone     string            `json:"zone"`
	Metadata map[string]string `json:"metadata"`
}

// PeerAddResp is the success response sent to a PeerAddReq request
type PeerAddResp Peer

// PeerEditResp is the success response sent to a PeerEditReq request
type PeerEditResp Peer

// PeerGetResp is the response sent for a peer get request
type PeerGetResp Peer

// PeerListResp is the response sent for a peer list request
/*
The client can request to filter peer listing based on metadata key/value using query parameters.
Example:
	- GET http://localhost:24007/v1/peers?key={keyname}&value={value}
	- GET http://localhost:24007/v1/peers?key={keyname}
	- GET http://localhost:24007/v1/peers?value={value}
Note - Cannot use query parameters if peerid is also supplied.
*/
type PeerListResp []PeerGetResp

// MetadataSize returns the size of the peer metadata in PeerAddReq
func (p *PeerAddReq) MetadataSize() int {
	return mapSize(p.Metadata)
}

// MetadataSize returns the size of the peer metadata in PeerEditReq
func (p *PeerEditReq) MetadataSize() int {
	return mapSize(p.Metadata)
}
