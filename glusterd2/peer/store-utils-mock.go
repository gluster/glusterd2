package peer

// This file contains mock functions to be used during unit tests.
// Do not use these functions in any other place.

import (
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/pborman/uuid"
)

// GetPeerIDByAddrMockGood returns a random peer ID when called
func GetPeerIDByAddrMockGood(addr string) (uuid.UUID, error) {
	return uuid.NewRandom(), nil
}

// GetPeerFMockGood returns a mock Peer
func GetPeerFMockGood(id string) (*Peer, error) {
	var p Peer
	p.Name = "test"
	p.ID = uuid.NewRandom()
	p.PeerAddresses = []string{"test"}
	return &p, nil
}

// GetPeerIDByAddrMockBad returns an error when called
func GetPeerIDByAddrMockBad(addr string) (uuid.UUID, error) {
	return nil, errors.ErrPeerNotFound
}
