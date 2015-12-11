// Package store implements the centralized store for GlusterD
//
// It currently uses [Consul](https://www.consul.io) as the backend. But we
// hope to have it pluggable in the end. libkv is being used to help achieve
// this goal.
package store

import (
	"time"

	log "github.com/Sirupsen/logrus"
	etcdctx "github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
)

const (
	// GlusterPrefix prefixes all paths in the store
	GlusterPrefix string = "gluster/"
)

var (
	prefixes []string
	EtcdCtx  etcdctx.Context
)

// GDStore is the GlusterD centralized store
type GDStore struct {
	client.KeysAPI
}

func init() {
	EtcdCtx = etcdctx.Background()
}

// New creates a new GDStore
func New() *GDStore {
	cfg := client.Config{
		//TODO: Make this configurable
		Endpoints: []string{"http://127.0.0.1:2379"},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
		log.WithField("error", err).Fatal("Failed to create store")
	}
	s := client.NewKeysAPI(c)

	log.Info("Created new store using etcd")

	gds := &GDStore{s}

	if e := gds.InitPrefix(GlusterPrefix); e != nil {
		log.Fatal("failed to init store prefixes")
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

	if _, e := s.Get(etcdctx.Background(), p, nil); e != nil {
		switch e.(client.Error).Code {
		case client.ErrorCodeKeyNotFound:
			log.WithField("prefix", p).Debug("prefix not found, initing")
			opts := new(client.SetOptions)
			opts.Dir = true
			if _, e := s.Set(EtcdCtx, p, "", opts); e != nil {
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
