package store

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
)

func newRemoteStore(conf *Config) (*GDStore, error) {
	var tlsConfig *tls.Config
	var c *clientv3.Client
	var e error
	if conf.UseTLS {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		if conf.ClntCertFile != "" && conf.ClntKeyFile != "" {
			tlsCert, err := tls.LoadX509KeyPair(conf.ClntCertFile, conf.ClntKeyFile)
			if err != nil {
				log.WithError(err).Error("failed to load client certificate file")
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{tlsCert}
			tlsConfig.ClientAuth = tls.RequestClientCert
		}
		if conf.ClntCAFile != "" {
			caCert, err := ioutil.ReadFile(conf.ClntCAFile)
			if err != nil {
				log.WithError(err).Error("failed to load client CA file")
				return nil, err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
	}
	c, e = clientv3.New(clientv3.Config{
		Endpoints:        conf.Endpoints,
		AutoSyncInterval: 30 * time.Second,
		DialTimeout:      5 * time.Second,
		RejectOldCluster: true,
		TLS:              tlsConfig,
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
