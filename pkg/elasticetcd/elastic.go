package elasticetcd

import (
	"io"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"

	"github.com/sirupsen/logrus"
)

// ElasticEtcd is an elastic etcd instance
type ElasticEtcd struct {
	cli    *clientv3.Client // the etcd client
	server                  // the embedded etcd server

	session  *concurrency.Session
	election *concurrency.Election

	conf *Config // the ElasticEtcd configuration

	log     *logrus.Logger
	logFile io.WriteCloser

	stopping bool

	stopwatching chan struct{}
	watchers     sync.WaitGroup

	lock sync.RWMutex
}

// New returns an initialized and connected ElasticEtcd, ready for use
func New(conf *Config) (*ElasticEtcd, error) {
	var serverStarted bool

	ee := new(ElasticEtcd)
	ee.conf = conf
	ee.stopwatching = make(chan struct{})
	ee.initLogging()

	// If no endpoints are given or if the default endpoint is set, assume that there is no existing server
	if len(ee.conf.Endpoints) == 0 || isDefaultEndpoint(ee.conf.Endpoints) {
		ee.log.Debug("no configured endpoints, starting own server")

		if err := ee.startServer(""); err != nil && err != ErrClientNotAvailable {
			ee.Stop()
			ee.log.WithError(err).Debug("failed to start server")
			return nil, err
		}

		// Update the endpoints to the advertised client urls of the embedded server
		ee.conf.Endpoints = ee.server.srv.Config().ACUrls
		serverStarted = true
	}

	// Connect the the etcd cluster as client first
	if err := ee.startClient(); err != nil {
		ee.Stop()
		return nil, err
	}

	if serverStarted {
		// Add yourself to the nominee list, avoids nominating yourself again when you become the leader
		ee.addToNominees(ee.conf.Name, ee.server.srv.Config().APUrls)
	}

	// Volunteer self and start watching for your nomination
	if err := ee.volunteerSelf(); err != nil {
		ee.Stop()
		return nil, err
	}

	// Start campaign to become the leader
	ee.startCampaign()

	return ee, nil
}

// Stop stops the ElasticEtcd instance
func (ee *ElasticEtcd) Stop() {
	ee.lock.Lock()
	defer ee.lock.Unlock()

	ee.stopping = true
	ee.stopClient()
	ee.stopServer()
	ee.logFile.Close()
}
