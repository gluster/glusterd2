package peer

// This file contains helper functions facilitate easier interaction with the
// peer information stores in the store

import (
	"encoding/json"

	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/store"

	log "github.com/Sirupsen/logrus"
)

const (
	peerPrefix string = store.GlusterPrefix + "peers/"
)

//func init() {
//context.Store.InitPrefix(peerPrefix)
//}

// AddOrUpdatePeer adds/updates given peer in the store
func AddOrUpdatePeer(p *Peer) error {
	json, err := json.Marshal(p)
	if err != nil {
		return err
	}

	idStr := p.ID.String()

	if err := context.Store.Put(peerPrefix+idStr, json, nil); err != nil {
		return err
	}

	return nil
}

// GetPeer returns specified peer from the store
func GetPeer(id string) (*Peer, error) {
	pair, err := context.Store.Get(peerPrefix + id)
	if err != nil || pair == nil {
		return nil, err
	}

	var p Peer
	if err := json.Unmarshal(pair.Value, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetPeers returns all available peers in the store
func GetPeers() ([]Peer, error) {
	pairs, err := context.Store.List(peerPrefix)
	if err != nil || pairs == nil {
		return nil, err
	}

	peers := make([]Peer, len(pairs))

	for i, pair := range pairs {
		var p Peer

		if err := json.Unmarshal(pair.Value, &p); err != nil {
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
func DeletePeer(id string) error {
	return context.Store.Delete(peerPrefix + id)
}

// Exists checks if given peer is present in the store
func Exists(id string) bool {
	b, e := context.Store.Exists(peerPrefix + id)
	if e != nil {
		return false
	}

	return b
}
