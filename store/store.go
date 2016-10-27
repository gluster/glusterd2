// Package store implements the centralized store for GlusterD
//
// We use etcd as the store backend, and use libkv as the frontend to etcd.
// libkv should allow us to change backends easily if required.
package store

import (
	"time"

	"github.com/gluster/glusterd2/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/etcd"
)

const (
	// GlusterPrefix prefixes all paths in the store
	GlusterPrefix string = "gluster/"
)

// GDStore is the GlusterD centralized store
type GDStore struct {
	store.Store
}

func init() {
	etcd.Register()
}

// New creates a new GDStore
func New() *GDStore {
	//TODO: Make this configurable
	ip, _ := utils.GetLocalIP()
	address := ip + ":2379"

	s, err := libkv.NewStore(store.ETCD, []string{address}, &store.Config{ConnectionTimeout: 10 * time.Second})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create libkv store.")
	}
	log.WithFields(log.Fields{"backend": "etcd", "client": address}).Debug("Created libkv store.")

	gds := &GDStore{s}
	return gds
}

// InitPrefix initializes the given prefix `p` in the store so that GETs on empty prefixes don't fail
// Returns error on failure, nil on success
func (s *GDStore) InitPrefix(p string) error {
	// Create the prefix if the prefix is not found. If any other error occurs
	// return it. Don't do anything if prefix is found
	if _, e := s.Get(p); e != nil {
		switch e {
		case store.ErrKeyNotFound:
			log.WithField("prefix", p).Debug("Prefix not found.")
			if e := s.Put(p, nil, &store.WriteOptions{IsDir: true}); e != nil {
				return e
			} else {
				log.WithField("prefix", p).Debug("Created prefix.")
			}

		default:
			return e
		}
	} else {
		log.WithField("prefix", p).Debug("Prefix found.")
	}

	return nil
}
