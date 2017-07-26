// Package store implements the centralized store for GlusterD
package store

import (
	"errors"
	"os"
	"sync"

	"github.com/gluster/glusterd2/pkg/elasticetcd"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

const (
	// GlusterPrefix prefixes all paths in the store
	GlusterPrefix = "gluster/"
	sessionTTL    = 30 // used for etcd mutexes and liveness key
)

var (
	// Store is the default GDStore that must to be used by the packages in GD2
	Store *GDStore
	lock  sync.Mutex

	// ErrEndpointsRequired is returned when endpoints are required but none are given
	ErrEndpointsRequired = errors.New("etcd endpoints for remote etcd cluster required with no-embed")
	// ErrStoreInitedAlready is returned when the store is already intialized
	ErrStoreInitedAlready = errors.New("store has been intialized already")
)

// GDStore is the GlusterD centralized store
type GDStore struct {
	conf Config

	*clientv3.Client
	*concurrency.Session

	ee *elasticetcd.ElasticEtcd
}

// Init initializes the GD2 store
func Init(conf *Config) error {
	lock.Lock()
	defer lock.Unlock()

	if Store != nil {
		return ErrStoreInitedAlready
	}

	var err error
	Store, err = New(conf)
	return err
}

// Close closes the GD2 store
func Close() {
	lock.Lock()
	defer lock.Unlock()

	Store.Close()
}

// Destroy closes the GD2 store and deletes the store data dir
func Destroy() {
	lock.Lock()
	defer lock.Unlock()

	Store.Destroy()
	Store = nil

	return
}

// New creates a new GDStore from the given Config.
// If the given Config is nil, the saved store config is used.
func New(conf *Config) (*GDStore, error) {
	if conf == nil {
		conf = GetConfig()
	}
	var (
		store *GDStore
		err   error
	)
	if conf.NoEmbed {
		if store, err = newRemoteStore(conf); err != nil {
			return nil, err
		}
	} else {
		if store, err = newEmbedStore(conf); err != nil {
			return nil, err
		}
	}

	if err = store.publishLiveness(); err != nil {
		return nil, err
	}

	return store, nil
}

// Close closes the store connections
func (s *GDStore) Close() {
	if s.ee != nil {
		s.closeEmbedStore()
	} else {
		s.closeRemoteStore()
	}
}

// Destroy closes the store and deletes the store data dir
func (s *GDStore) Destroy() {
	s.Close()
	os.RemoveAll(s.conf.Dir)
}

// UpdateEndpoints updates the configured endpoints and saves them
func (s *GDStore) UpdateEndpoints() error {
	if err := s.Sync(s.Ctx()); err != nil {
		return err
	}

	s.conf.Endpoints = s.Client.Endpoints()
	return s.conf.Save()
}
