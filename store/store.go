// Package store implements the centralized store for GlusterD
package store

import (
	"context"
	"time"

	"github.com/gluster/glusterd2/gdctx"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/pborman/uuid"
	config "github.com/spf13/viper"
)

const (
	// GlusterPrefix prefixes all paths in the store
	GlusterPrefix     = "gluster/"
	sessionTTL        = 30 // used for etcd mutexes and liveness key
	livenessKeyPrefix = GlusterPrefix + "alive/"
)

// Store variable can be imported by packages which need access to the store
var Store *GDStore

// GDStore is the GlusterD centralized store
type GDStore struct {
	*clientv3.Client
	*concurrency.Session
}

// InitStore creates and initializes the store
func InitStore() error {
	address := config.GetString("etcdclientaddress")

	c, e := clientv3.New(clientv3.Config{
		Endpoints:        []string{address},
		AutoSyncInterval: 1 * time.Minute,
		DialTimeout:      10 * time.Second,
	})
	if e != nil {
		log.WithError(e).Error("failed to create etcd client")
		return e
	}
	log.Debug("etcd client connection created")

	// Create a new session (lease kept alive for the lifetime of a client)
	// This is currently used for:
	// * distributed locking (Mutex)
	// * representing liveness of the client
	s, e := concurrency.NewSession(c, concurrency.WithTTL(sessionTTL))
	if e != nil {
		log.WithError(e).Error("failed to create an etcd session")
		return e
	}

	// publish liveness of this instance into the store
	key := livenessKeyPrefix + gdctx.MyUUID.String()
	if _, e = c.Put(context.TODO(), key, "", clientv3.WithLease(s.Lease())); e != nil {
		return e
	}

	Store = &GDStore{c, s}
	return nil
}

// Close closes the store connections
func (s *GDStore) Close() {
	s.Session.Orphan()
	if e := s.Client.Close(); e != nil {
		log.WithError(e).Warn("failed to close etcd client connection")
	}
	// FIXME: We should close the session first and then the client but it
	// doesn't work because restart of embedded etcd server when using v3
	// has issues.
	if e := s.Session.Close(); e != nil {
		log.WithError(e).Warn("failed to close etcd session")
	}
}

// IsNodeAlive returns true if the node specified is alive as seen by the store
func (s *GDStore) IsNodeAlive(nodeID interface{}) bool {

	var keySuffix string

	switch nodeID.(type) {
	case uuid.UUID:
		keySuffix = nodeID.(uuid.UUID).String()
	case string:
		keySuffix = nodeID.(string)
		if uuid.Parse(keySuffix) == nil {
			return false
		}
	default:
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := livenessKeyPrefix + keySuffix
	resp, err := s.Client.Get(ctx, key)
	if err != nil {
		return false
	}

	return resp.Count == 1
}
