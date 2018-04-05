package api

import "github.com/pborman/uuid"

// Peer represents a peer in the glusterd2 cluster
type Peer struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	PeerAddresses   []string          `json:"peer-addresses"`
	ClientAddresses []string          `json:"client-addresses"`
	Online          bool              `json:"online"`
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
type PeerListResp []PeerGetResp
