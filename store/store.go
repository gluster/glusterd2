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

	// Create a new locking session to be used for locking in transaction and other places
	s, e := concurrency.NewSession(c)
	if e != nil {
		log.WithError(e).Error("failed to create an etcd session")
		return e
	}

	Store = &GDStore{c, s}
	return nil
}

// Close closes the store connections
func (s *GDStore) Close() {
	if e := s.Client.Close(); e != nil {
		log.WithError(e).Warn("failed to close etcd client connection")
	}
	if e := s.Session.Close(); e != nil {
		log.WithError(e).Warn("failed to close etcd session")
	}
}
