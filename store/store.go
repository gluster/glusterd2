// Package store implements the centralized store for GlusterD
//
// It currently uses [Consul](https://www.consul.io) as the backend. But we
// hope to have it pluggable in the end. libkv is being used to help achieve
// this goal.
package store

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
)

const (
	glusterPrefix string = "gluster/"
)

var (
	prefixes []string
)

// GDStore is the GlusterD centralized store
type GDStore struct {
	store.Store
}

func init() {
	consul.Register()
	prefixes = append(prefixes, glusterPrefix)
}

// New creates a new GDStore
func New() *GDStore {
	//TODO: Make this configurable
	address := "localhost:8500"
	consul.Register()

	log.WithFields(log.Fields{"type": "consul", "consul.config": address}).Debug("Creating new store")
	s, err := libkv.NewStore(store.CONSUL, []string{address}, &store.Config{ConnectionTimeout: 10 * time.Second})
	if err != nil {
		log.WithField("error", err).Fatal("Failed to create store")
	}

	log.Info("Created new store using Consul")

	gds := &GDStore{s}

	if !gds.initPrefixes() {
		log.Fatal("failed to init store prefixes")
	}

	return gds
}

// initPrefixes initalizes the store prefixes so that GETs on empty prefixes don't fail
// Returns true on success, false otherwise.
func (s *GDStore) initPrefixes() bool {
	log.Debug("initing store prefixes")
	for _, p := range prefixes {
		// Create the prefix if the prefix is not found. If any other error occurs
		// return it. Don't do anything if prefix is found
		if _, e := s.Get(p); e != nil {
			switch e {
			case store.ErrKeyNotFound:
				log.WithField("prefix", p).Debug("prefix not found, initing")

				if e := s.Put(p, nil, nil); e != nil {
					log.WithFields(log.Fields{
						"preifx": p,
						"error":  e,
					}).Error("error initing prefix")

					return false
				}

			default:
				log.WithFields(log.Fields{
					"prefix": p,
					"error":  e,
				}).Error("error getting prefix")

				return false
			}
		} else {
			log.WithField("prefix", p).Debug("prefix present")
		}
	}
	return true
}
