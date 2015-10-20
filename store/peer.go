package store

// This file contains helper functions facilitate easier interaction with the
// peer information stores in the store

import (
	"encoding/json"

	"github.com/gluster/glusterd2/peer"

	log "github.com/Sirupsen/logrus"
)

const (
	peerPrefix string = glusterPrefix + "peers/"
)

// AddOrUpdatePeer adds/updates given peer in the store
func (s *GDStore) AddOrUpdatePeer(p *peer.Peer) error {
	json, err := json.Marshal(p)
	if err != nil {
		return err
	}

	if err := s.Put(peerPrefix+p.IDstr, json, nil); err != nil {
		return err
	}

	return nil
}

// GetPeer returns specified peer from the store
func (s *GDStore) GetPeer(id string) (*peer.Peer, error) {
	pair, err := s.Get(peerPrefix + id)
	if err != nil || pair == nil {
		return nil, err
	}

	return peer.UnmarshalJSON(pair.Value)
}

// GetPeers returns all available peers in the store
func (s *GDStore) GetPeers() ([]peer.Peer, error) {
	pairs, err := s.List(peerPrefix)
	if err != nil || pairs == nil {
		return nil, err
	}

	peers := make([]peer.Peer, len(pairs))

	for i, pair := range pairs {
		var p peer.Peer

		if err := p.UnmarshalJSON(pair.Value); err != nil {
			log.WithFields(log.Fields{
				"peer":  pair.Key,
				"error": err,
			}).Error("Failed to unmarshal peer")
			continue
		}
		peers[i] = p
	}

	return peers, nil
}

// DeletePeer deletes given peer from the store
func (s *GDStore) DeletePeer(id string) error {
	return s.Delete(peerPrefix + id)
}

// PeerExists checks if given peer is present in the store
func (s *GDStore) PeerExists(id string) (bool, error) {
	return s.Exists(peerPrefix + id)
}
