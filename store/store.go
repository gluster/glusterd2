// Package store implements the centralized store for GlusterD
package store

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	config "github.com/spf13/viper"
)

const (
	// GlusterPrefix prefixes all paths in the store
	GlusterPrefix string = "gluster/"
)

// GDStore is the GlusterD centralized store
type GDStore struct {
	*clientv3.Client
	*concurrency.Session
}

// New creates a new GDStore
func New() *GDStore {
	address := config.GetString("etcdclientaddress")

	c, e := clientv3.New(clientv3.Config{
		Endpoints:        []string{address},
		AutoSyncInterval: 1 * time.Minute,
		DialTimeout:      10 * time.Second,
	})
	if e != nil {
		log.WithError(e).Fatal("failed to create etcd client")
		return nil
	}
	log.Debug("etcd client connection created")

	// Create a new locking session to be used for locking in transaction and other places
	s, e := concurrency.NewSession(c)
	if e != nil {
		log.WithError(e).Fatal("failed to create an etcd session")
		return nil
	}

	return &GDStore{c, s}
}

// Close closes the store connections
func (s *GDStore) Close() {
	if e := s.Session.Close(); e != nil {
		log.WithError(e).Warn("failed to close etcd session")
	}
	if e := s.Client.Close(); e != nil {
		log.WithError(e).Warn("failed to close etcd client connection")
	}
}
