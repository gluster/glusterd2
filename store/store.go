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
	// Store variable can be imported by packages which need access to the store
	Store *GDStore
	lock  sync.Mutex

	ErrEndpointsRequired  = errors.New("etcd endpoints for remote etcd cluster required with no-embed")
	ErrStoreInitedAlready = errors.New("store has been intialized already")
)

// GDStore is the GlusterD centralized store
type GDStore struct {
	conf *Config
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

// New creates a new GDStore
func New(conf *Config) (*GDStore, error) {
	if conf == nil {
		conf = getConf()
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
