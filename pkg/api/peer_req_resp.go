package api

import "github.com/pborman/uuid"

// Peer represents a peer in the glusterd2 cluster
type Peer struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Addresses []string  `json:"addresses"`
	Online    bool      `json:"online"`
	Group     int       `json:"group"`
}

// PeerAddReq represents an incoming request to add a peer to the cluster
type PeerAddReq struct {
	Addresses []string `json:"addresses"`
	Group     int      `json:"group"`
}

// PeerAddResp is the success response sent to a PeerAddReq request
type PeerAddResp Peer

// PeerGetResp is the response sent for a peer get request
type PeerGetResp Peer

// PeerListResp is the response sent for a peer list request
type PeerListResp []PeerGetResp
