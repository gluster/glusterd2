package store

import (
	"time"

	"github.com/gluster/glusterd2/pkg/elasticetcd"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

func newRemoteStore(conf *Config) (*GDStore, error) {
	if len(conf.Endpoints) == 1 && conf.Endpoints[1] == elasticetcd.DefaultEndpoint {
		return nil, ErrEndpointsRequired
	}

	c, e := clientv3.New(clientv3.Config{
		Endpoints:        conf.Endpoints,
		AutoSyncInterval: 1 * time.Minute,
		DialTimeout:      10 * time.Second,
	})
	if e != nil {
		log.WithError(e).Error("failed to create etcd client")
		return nil, e
	}
	log.Debug("etcd client connection created")

	// Create a new session (lease kept alive for the lifetime of a client)
	// This is currently used for:
	// * distributed locking (Mutex)
	// * representing liveness of the client
	s, e := concurrency.NewSession(c, concurrency.WithTTL(sessionTTL))
	if e != nil {
		log.WithError(e).Error("failed to create an etcd session")
		return nil, e
	}

	return &GDStore{conf, c, s, nil}, nil
}

func (s *GDStore) closeRemoteStore() {
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
