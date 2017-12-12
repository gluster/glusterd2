package store

import (
	"time"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
)

func newRemoteStore(conf *Config) (*GDStore, error) {

	c, e := clientv3.New(clientv3.Config{
		Endpoints:        conf.Endpoints,
		AutoSyncInterval: 30 * time.Second,
		DialTimeout:      5 * time.Second,
		RejectOldCluster: true,
	})
	if e != nil {
		log.WithError(e).Error("failed to create etcd client")
		return nil, e
	}
	log.Debug("etcd client connection created")

	return newNamespacedStore(c, conf)
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
