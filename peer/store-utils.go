package peer

// This file contains helper functions facilitate easier interaction with the
// peer information stores in the store

import (
	"encoding/json"

	"github.com/gluster/glusterd2/errors"
	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/store"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

const (
	peerPrefix string = store.GlusterPrefix + "peers/"
)

var (
	GetPeerF         = GetPeer
	GetPeersF        = GetPeers
	GetPeerByAddrF   = GetPeerByAddr
	GetPeerByNameF   = GetPeerByName
	GetPeerIDByAddrF = GetPeerIDByAddr
)

func init() {
	gdctx.RegisterStorePrefix(peerPrefix)
}

// AddOrUpdatePeer adds/updates given peer in the store
func AddOrUpdatePeer(p *Peer) error {
	json, err := json.Marshal(p)
	if err != nil {
		return err
	}

	idStr := p.ID.String()

	if err := gdctx.Store.Put(peerPrefix+idStr, json, nil); err != nil {
		return err
	}

	return nil
}

// GetPeer returns specified peer from the store
func GetPeer(id string) (*Peer, error) {
	pair, err := gdctx.Store.Get(peerPrefix + id)
	if err != nil || pair == nil {
		return nil, err
	}

	var p Peer
	if err := json.Unmarshal(pair.Value, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetInitialCluster() form and returns the etcd initial cluster value in a
// string
func GetInitialCluster() (string, error) {
	var initialCluster string
	peers, err := GetPeersF()
	if err != nil {
		return "", err
	}
	c := 0
	for _, peer := range peers {
		if peer.Client == true {
			continue
		}
		if c > 0 {
			initialCluster = initialCluster + ", "
		}
		initialCluster = initialCluster + peer.Name + "=" + "http://" + peer.Name + ":2380"
		c = c + 1
	}
	return initialCluster, nil
}

// GetPeers returns all available peers in the store
func GetPeers() ([]Peer, error) {
	pairs, err := gdctx.Store.List(peerPrefix)
	if err != nil || pairs == nil {
		return nil, err
	}
	// There will be at least one peer (current node)
	peers := make([]Peer, len(pairs))
	i := 0
	for _, pair := range pairs {
		var p Peer

		if err := json.Unmarshal(pair.Value, &p); err != nil {
			log.WithFields(log.Fields{
				"peer":  pair.Key,
				"error": err,
			}).Error("Failed to unmarshal peer")
			continue
		}
		peers[i] = p
		i = i + 1
	}

	return peers, nil
}

// GetPeerByName returns the peer with the given name from store
func GetPeerByName(name string) (*Peer, error) {
	pairs, err := gdctx.Store.List(peerPrefix)
	if err != nil || pairs == nil {
		return nil, err
	}

	for _, pair := range pairs {
		var p Peer
		if err := json.Unmarshal(pair.Value, &p); err != nil {
			log.WithFields(log.Fields{
				"peer":  pair.Key,
				"error": err,
			}).Error("Failed to unmarshal peer")
			continue
		}
		if p.Name == name {
			return &p, nil
		}
	}

	return nil, errors.ErrPeerNotFound
}

// DeletePeer deletes given peer from the store
func DeletePeer(id string) error {
	return gdctx.Store.Delete(peerPrefix + id)
}

// Exists checks if given peer is present in the store
func Exists(id string) bool {
	b, e := gdctx.Store.Exists(peerPrefix + id)
	if e != nil {
		return false
	}

	return b
}

//GetPeerByAddr returns the peer with the given address from the store
func GetPeerByAddr(addr string) (*Peer, error) {
	pairs, err := gdctx.Store.List(peerPrefix)
	if err != nil || pairs == nil {
		return nil, err
	}

	for _, pair := range pairs {
		var p Peer
		if err := json.Unmarshal(pair.Value, &p); err != nil {
			log.WithFields(log.Fields{
				"peer":  pair.Key,
				"error": err,
			}).Error("Failed to unmarshal peer")
			continue
		}

		for _, paddr := range p.Addresses {
			if paddr == addr {
				return &p, nil
			}
		}
	}

	return nil, errors.ErrPeerNotFound
}

//GetPeerIDByAddr returns the ID of the peer with the given address
func GetPeerIDByAddr(addr string) (uuid.UUID, error) {
	p, e := GetPeerByAddrF(addr)
	if e != nil {
		return nil, errors.ErrPeerNotFound
	} else {
		return p.ID, nil
	}

}
