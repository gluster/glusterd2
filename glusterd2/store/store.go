// Package store implements the centralized store for GlusterD
package store

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/elasticetcd"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/clientv3/namespace"
	log "github.com/sirupsen/logrus"
)

const (
        GlusterPrefix = "gluster/"
	sessionTTL = 30 // used for etcd mutexes and liveness key
)

var (
	// Store is the default GDStore that must to be used by the packages in GD2
	Store *GDStore
	lock  sync.Mutex

	// ErrStoreInitedAlready is returned when the store is already intialized
	ErrStoreInitedAlready = errors.New("store has been intialized already")
)

// GDStore is the GlusterD centralized store
type GDStore struct {
	conf Config

	// Namespaced KV, Lease and Watchers used for performing namespaced operations in the Store
	clientv3.KV
	clientv3.Lease
	clientv3.Watcher
	// Namespaced Session for session and concurrency operations in the Store
	*concurrency.Session
	// Un-namespaced Client for Auth, Cluster and Maintenance operations
	*clientv3.Client

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

// Returns a new GDStore from the given etcd Client, with namespaced KV, Lease, Watcher, Session etc.
func newNamespacedStore(oc *clientv3.Client, conf *Config) (*GDStore, error) {
	namespaceKey := fmt.Sprintf("gluster-%s/", gdctx.MyClusterID.String())

	// Create namespaced interfaces
	kv := namespace.NewKV(oc.KV, namespaceKey)
	lease := namespace.NewLease(oc.Lease, namespaceKey)
	watcher := namespace.NewWatcher(oc.Watcher, namespaceKey)

	// Create a new session (lease kept alive for the lifetime of a client)
	// This is currently used for:
	// * distributed locking (Mutex)
	// * representing liveness of the client

	// Creating a client with the namespaced variants KV, Lease and Watcher for creating a namespaced Session
	nc := clientv3.NewCtxClient(oc.Ctx())
	nc.KV = kv
	nc.Lease = lease
	nc.Watcher = watcher

	// Create new Session from the new namespaced client
	session, err := concurrency.NewSession(nc, concurrency.WithTTL(sessionTTL))
	if err != nil {
		log.WithError(err).Error("failed to create an etcd session")
		return nil, err
	}

	return &GDStore{*conf, kv, lease, watcher, session, oc, nil}, nil
}
