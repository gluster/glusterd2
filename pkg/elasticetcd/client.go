package elasticetcd

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	log "github.com/sirupsen/logrus"
)

// Client returns the etcd client of ElasticEtcd
func (ee *ElasticEtcd) Client() *clientv3.Client {
	ee.lock.RLock()
	defer ee.lock.RUnlock()
	return ee.cli
}

// Session returns the etcd session used by ElasticEtcd
func (ee *ElasticEtcd) Session() *concurrency.Session {
	ee.lock.RLock()
	defer ee.lock.RUnlock()
	return ee.session
}

// startClient starts the etcd client and connects the ElasticEtcd instance to the elastic cluster.
func (ee *ElasticEtcd) startClient() error {
	if ee.cli != nil {
		return errors.New("client already exists")
	}

	cli, err := clientv3.New(ee.newClientConfig())
	if err != nil {
		return err
	}

	ee.cli = cli
	// Immediately sync and update your list of endpoints
	ee.cli.Sync(ee.cli.Ctx())

	// Begin a new session, which is needed for the watchers
	session, err := concurrency.NewSession(ee.cli)
	if err != nil {
		ee.cli.Close()
		return err
	}
	ee.session = session

	return nil
}

func (ee *ElasticEtcd) stopClient() error {
	if ee.cli == nil {
		return errors.New("no client present")
	}

	// First stop all the watchers
	close(ee.stopwatching)
	ee.watchers.Wait()

	// Then close the session
	if err := ee.session.Close(); err != nil {
		return err
	}

	// Then close the etcd client
	if err := ee.cli.Close(); err != nil {
		return err
	}

	ee.cli = nil
	return nil
}

// newClientConfig returns a new etcd clientv3.Config from the ElasticEtcd config
func (ee *ElasticEtcd) newClientConfig() clientv3.Config {
	var tlsConfig *tls.Config
	if ee.conf.UseTLS {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		if ee.conf.ClntCertFile != "" && ee.conf.ClntKeyFile != "" {
			tlsCert, err := tls.LoadX509KeyPair(ee.conf.ClntCertFile, ee.conf.ClntKeyFile)
			if err != nil {
				log.WithError(err).Error("failed to load certificate file")
				return clientv3.Config{}
			}
			tlsConfig.Certificates = []tls.Certificate{tlsCert}
			tlsConfig.ClientAuth = tls.RequestClientCert
		}
		if ee.conf.CAFile != "" {
			caCert, err := ioutil.ReadFile(ee.conf.CAFile)
			if err != nil {
				log.WithError(err).Error("failed to load CA file")
				return clientv3.Config{}
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
	}

	return clientv3.Config{
		Endpoints:        ee.conf.Endpoints.StringSlice(),
		AutoSyncInterval: 30 * time.Second, // Update list of endpoints ever 30s.
		DialTimeout:      5 * time.Second,
		RejectOldCluster: true,
		TLS:              tlsConfig,
	}
}

// watch watches for changes the given key and runs the handler when changes happen.
// watch also waits on the stopwatching channel and stops watching when notified.
// All watchers in ElasticEtcd must to use this instead of using starting their own etcd watchers.
func (ee *ElasticEtcd) watch(key string, handler func(clientv3.WatchResponse), watchopts ...clientv3.OpOption) {
	ee.watchers.Add(1)
	go func() {
		defer ee.watchers.Done()

		wch := ee.cli.Watch(ee.cli.Ctx(), key, watchopts...)
		for {
			select {
			case resp := <-wch:
				if resp.Canceled {
					return
				}
				handler(resp)
			case <-ee.stopwatching:
				return
			}
		}
	}()
}
