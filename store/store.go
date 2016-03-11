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

var (
	prefixes []string
)

// GDStore is the GlusterD centralized store
type GDStore struct {
	store.Store
}

func init() {
	etcd.Register()
}

// New creates a new GDStore
func New(restart bool) *GDStore {
	//TODO: Make this configurable
	ip, _ := utils.GetLocalIP()
	address := ip + ":2379"
	log.WithFields(log.Fields{"type": "etcd", "etcd.config": address}).Debug("Creating new store")
	s, err := libkv.NewStore(store.ETCD, []string{address}, &store.Config{ConnectionTimeout: 10 * time.Second})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create store")
	}

	log.Info("Created new store using ETCD")

	gds := &GDStore{s}

	if restart == false {
		if e := gds.InitPrefix(GlusterPrefix); e != nil {
			log.Fatal("failed to init store prefixes")
		}
	}

	return gds
}

// initPrefixes initalizes the store prefixes so that GETs on empty prefixes don't fail
// Returns true on success, false otherwise.
func (s *GDStore) initPrefixes() bool {
	log.Debug("initing store prefixes")
	for _, p := range prefixes {
		if e := s.InitPrefix(p); e != nil {
			return false
		}
	}
	return true
}

// InitPrefix initializes the given prefix `p` in the store so that GETs on empty prefixes don't fail
// Returns error on failure, nil on success
func (s *GDStore) InitPrefix(p string) error {
	// Create the prefix if the prefix is not found. If any other error occurs
	// return it. Don't do anything if prefix is found
	if _, e := s.Get(p); e != nil {
		switch e {
		case store.ErrKeyNotFound:
			log.WithField("prefix", p).Debug("prefix not found, initing")

			if e := s.Put(p, nil, &store.WriteOptions{IsDir: true}); e != nil {
				log.WithFields(log.Fields{
					"preifx": p,
					"error":  e,
				}).Error("error initing prefix")

				return e
			}

		default:
			log.WithFields(log.Fields{
				"prefix": p,
				"error":  e,
			}).Error("error getting prefix")

			return e
		}
	} else {
		log.WithField("prefix", p).Debug("prefix present")
	}

	return nil
}
