package peer

// This file contains helper functions facilitate easier interaction with the
// peer information stores in the store

import (
	"context"
	"encoding/json"

	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/errors"
	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	peerPrefix string = "peers/"
)

var (
	// GetPeerF returns specified peer from the store
	GetPeerF = GetPeer
	// GetPeersF returns all available peers in the store
	GetPeersF = GetPeers
	//GetPeerByAddrF returns the peer with the given address from the store
	GetPeerByAddrF = GetPeerByAddr
	// GetPeerByNameF returns the peer with the given name from store
	GetPeerByNameF = GetPeerByName
	//GetPeerIDByAddrF returns the ID of the peer with the given address
	GetPeerIDByAddrF = GetPeerIDByAddr
)

// AddOrUpdatePeer adds/updates given peer in the store
func AddOrUpdatePeer(p *Peer) error {
	json, err := json.Marshal(p)
	if err != nil {
		return err
	}

	idStr := p.ID.String()

	if _, err := store.Store.Put(context.TODO(), peerPrefix+idStr, string(json)); err != nil {
		return err
	}

	return nil
}

// GetPeer returns specified peer from the store
func GetPeer(id string) (*Peer, error) {
	resp, err := store.Store.Get(context.TODO(), peerPrefix+id)
	if err != nil {
		return nil, err
	}

	// We cannot have more than one peer with a given ID
	// TODO: Fix this to return a proper error
	if resp.Count != 1 {
		return nil, errors.ErrPeerNotFound
	}

	var p Peer
	if err := json.Unmarshal(resp.Kvs[0].Value, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetInitialCluster forms and returns the etcd initial cluster value as a string
func GetInitialCluster() (string, error) {
	var initialCluster string
	peers, err := GetPeersF()
	if err != nil {
		return "", err
	}
	c := 0
	for _, peer := range peers {
		if c > 0 {
			initialCluster = initialCluster + ", "
		}
		initialCluster = initialCluster + peer.Name + "=" + "http://" + peer.Name + ":2380"
		c = c + 1
	}
	return initialCluster, nil
}

// GetPeers returns all available peers in the store
func GetPeers() ([]*Peer, error) {
	resp, err := store.Store.Get(context.TODO(), peerPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	// There will be at least one peer (current node)
	peers := make([]*Peer, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		var p Peer

		if err := json.Unmarshal(kv.Value, &p); err != nil {
			log.WithFields(log.Fields{
				"peer":  string(kv.Key),
				"error": err,
			}).Error("Failed to unmarshal peer")
			continue
		}
		peers[i] = &p
	}

	return peers, nil
}

// GetPeerIDs returns peer id (uuid) of all peers in the store
func GetPeerIDs() ([]uuid.UUID, error) {
	resp, err := store.Store.Get(context.TODO(), peerPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	uuids := make([]uuid.UUID, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		var p Peer
		if err := json.Unmarshal(kv.Value, &p); err != nil {
			log.WithFields(log.Fields{
				"peer":  string(kv.Key),
				"error": err,
			}).Error("Failed to unmarshal peer")
			continue
		}
		uuids[i] = p.ID
	}

	return uuids, nil
}

// GetPeerByName returns the peer with the given name from store
func GetPeerByName(name string) (*Peer, error) {
	peers, err := GetPeers()
	if err != nil {
		return nil, err
	}

	for _, p := range peers {
		if p.Name == name {
			return p, nil
		}
	}

	return nil, errors.ErrPeerNotFound
}

// DeletePeer deletes given peer from the store
func DeletePeer(id string) error {
	_, e := store.Store.Delete(context.TODO(), peerPrefix+id)
	return e
}

// Exists checks if given peer is present in the store
func Exists(id string) bool {
	resp, e := store.Store.Get(context.TODO(), peerPrefix+id)
	if e != nil {
		return false
	}

	return resp.Count == 1
}

// GetPeerByAddr returns the peer with the given address from the store
func GetPeerByAddr(addr string) (*Peer, error) {
	peers, e := GetPeers()
	if e != nil {
		return nil, e
	}

	for _, p := range peers {
		for _, paddr := range p.Addresses {
			if utils.IsPeerAddressSame(addr, paddr) {
				return p, nil
			}
		}
	}

	return nil, errors.ErrPeerNotFound
}

// GetPeerByAddrs returns a peer that matches any one of the given addresses
func GetPeerByAddrs(addrs []string) (*Peer, error) {
	peers, err := GetPeers()
	if err != nil {
		return nil, err
	}
	for _, a := range addrs {
		for _, p := range peers {
			for _, paddr := range p.Addresses {
				if utils.IsPeerAddressSame(a, paddr) {
					return p, nil
				}
			}
		}
	}
	return nil, errors.ErrPeerNotFound
}

// GetPeerIDByAddr returns the ID of the peer with the given address
func GetPeerIDByAddr(addr string) (uuid.UUID, error) {
	p, e := GetPeerByAddrF(addr)
	if e != nil {
		return nil, errors.ErrPeerNotFound
	}
	return p.ID, nil
}
