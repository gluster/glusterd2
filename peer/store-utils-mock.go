package peer

// This file contains mock functions to be used during unit tests.
// Do not use these functions in any other place.

import (
	"github.com/gluster/glusterd2/errors"
	"github.com/pborman/uuid"
)

// GetPeerIDByAddrMockGood returns a random peer ID when called
func GetPeerIDByAddrMockGood(addr string) (uuid.UUID, error) {
	return uuid.NewRandom(), nil
}

// GetPeerIDByAddrMockBad returns an error when called
func GetPeerIDByAddrMockBad(addr string) (uuid.UUID, error) {
	return nil, errors.ErrPeerNotFound
}
