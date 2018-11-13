// Package store implements the centralized store for GlusterD
package store

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/elasticetcd"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/clientv3/namespace"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

const (
	sessionTTL    = 30 // used for etcd mutexes and liveness key
	getTimeout    = 5
	putTimeout    = 5
	deleteTimeout = 5
)

var storeCounters = expvar.NewMap("store")

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

	ee        *elasticetcd.ElasticEtcd
	namespace string
	stop      chan struct{}
	stopOnce  sync.Once
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
func Destroy(deleteNamespace bool) {
	lock.Lock()
	defer lock.Unlock()

	Store.Destroy(deleteNamespace)
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

	store.stop = make(chan struct{})

	go store.keepSessionLeaseAlive()
	return store, nil
}

// keepSessionLeaseAlive keeps session lease alive forever. If etcd goes down
// and comes up then existing session lease will not be refreshed and expires after
// lease ttl duration. keepSessionLeaseAlive checks session lease information on a
// regular interval. It will start refreshing session lease if it is not getting
// refreshed.
func (s *GDStore) keepSessionLeaseAlive() {
	var (
		ticker         = time.NewTicker(time.Second * 5)
		printedFailure bool
	)

	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), getTimeout*time.Second)
			leaseTTLResp, err := s.Lease.TimeToLive(ctx, s.Session.Lease())
			cancel()

			if err != nil && !printedFailure {
				log.WithError(err).Error("etcd is probably down, restart etcd")
				printedFailure = true
			}

			if err != nil || leaseTTLResp.TTL > sessionTTL/3 {
				continue
			}

			// session lease ttl is not getting refreshed, keeping alive session lease.
			log.WithField("leaseID", leaseTTLResp.ID).
				Debugf("granted session lease will expire in %d sec, etcd might have restarted", leaseTTLResp.TTL)

			keepAliveChan, err := s.Lease.KeepAlive(context.Background(), s.Session.Lease())
			if err != nil || keepAliveChan == nil {
				continue
			}

			log.WithField("leaseID", leaseTTLResp.ID).Debug("keeping session lease alive forever")

			go func() {
				for range keepAliveChan {
					// discard messages until keepAliveChan closes
				}
			}()
			printedFailure = false
		}
	}
}

// Close closes the store connections
func (s *GDStore) Close() {
	if err := s.revokeLiveness(); err != nil {
		log.WithError(err).Error("failed to revoke liveness")
	}

	if s.ee != nil {
		s.closeEmbedStore()
	} else {
		s.closeRemoteStore()
	}

	s.stopOnce.Do(func() {
		close(s.stop)
	})
}

// Destroy closes the store and deletes the store data dir
func (s *GDStore) Destroy(deleteNamespace bool) {
	if s.ee != nil {
		s.Close()
		os.RemoveAll(s.conf.Dir)
		return
	}

	// remote store: delete the current namespace using un-namespaced
	// client and then close the client.
	if deleteNamespace {
		_, err := s.Client.Delete(context.Background(), s.namespace, clientv3.WithPrefix())
		if err != nil {
			log.WithError(err).Error("failed to delete etcd namespace during remote store destroy")
		}
	}
	s.Close()
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

	return &GDStore{
		conf:      *conf,
		KV:        kv,
		Lease:     lease,
		Watcher:   watcher,
		Session:   session,
		Client:    oc,
		ee:        nil,
		namespace: namespaceKey,
	}, nil
}

//Get is a wrapper function that calls clientv3.KV.Get with a default timeout if an empty context is passed
func Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	var cancel context.CancelFunc

	if ctx == context.TODO() {
		ctx, cancel = context.WithTimeout(context.Background(), getTimeout*time.Second)
		defer cancel()
	} else {
		var span *trace.Span
		ctx, span = trace.StartSpan(ctx, "store.Get")
		defer span.End()
	}

	defer storeCounters.Add("get", 1)
	return Store.Get(ctx, key, opts...)
}

//Put is a wrapper function that calls clientv3.KV.Put with a default timeout if an empty context is passed
func Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	var cancel context.CancelFunc

	if ctx == context.TODO() {
		ctx, cancel = context.WithTimeout(context.Background(), putTimeout*time.Second)
		defer cancel()
	}

	defer storeCounters.Add("put", 1)
	return Store.Put(ctx, key, val, opts...)
}

//Delete is a wrapper function that calls clientv3.KV.Delete with a default timeout if an empty context is passed
func Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	var cancel context.CancelFunc

	if ctx == context.TODO() {
		ctx, cancel = context.WithTimeout(context.Background(), deleteTimeout*time.Second)
		defer cancel()
	}

	defer storeCounters.Add("delete", 1)
	return Store.Delete(ctx, key, opts...)
}

// Txn is a wrapper function that calls clientv3.KV.Txn which creates a transaction
func Txn(ctx context.Context) clientv3.Txn {
	// can't cancel() here as caller will have to eventually call
	// clientv3.Txn.Commit()
	defer storeCounters.Add("txn", 1)
	return Store.Txn(ctx)
}
