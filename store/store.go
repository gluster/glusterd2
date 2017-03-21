// Package store implements the centralized store for GlusterD
//
// We use etcd as the store backend, and use libkv as the frontend to etcd.
// libkv should allow us to change backends easily if required.
package store

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	config "github.com/spf13/viper"
)

const (
	// GlusterPrefix prefixes all paths in the store
	GlusterPrefix string = "gluster/"
	directoryVal         = "thisisadirectory"
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

//// InitPrefix initializes the given prefix `p` in the store so that GETs on empty prefixes don't fail
//// Returns error on failure, nil on success
func (s *GDStore) InitPrefix(p string) error {
	// Create the prefix if the prefix is not found.
	// Don't do anything if prefix is found
	_, e := s.KV.Txn(context.TODO()).
		If(clientv3.Compare(clientv3.Version(p), "=", 0)).
		Then(clientv3.OpPut(p, directoryVal)).
		Commit()

	return e
}
