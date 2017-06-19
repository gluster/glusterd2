package elasticetcd

import (
	"io"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"

	"github.com/Sirupsen/logrus"
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

	stopwatching chan bool
	watchers     sync.WaitGroup

	lock sync.RWMutex
}

// New returns an initialized and connected ElasticEtcd, ready for use
func New(conf *Config) (*ElasticEtcd, error) {
	var serverStarted bool

	ee := new(ElasticEtcd)
	ee.conf = conf
	ee.stopwatching = make(chan bool)
	ee.initLogging()

	if len(ee.conf.Endpoints) == 0 || isDefaultEndpoint(ee.conf.Endpoints) {
		ee.log.Debug("no configured endpoints, starting own server")

		if err := ee.startServer(""); err != nil && err != ErrClientNotAvailable {
			ee.log.WithError(err).Debug("failed to start server")
			return nil, err
		}

		ee.conf.Endpoints = ee.server.srv.Config().ACUrls
		serverStarted = true
	}

	if err := ee.startClient(); err != nil {
		ee.Stop()
		return nil, err
	}

	if serverStarted {
		// Add yourself to the nominee list
		ee.addToNominees(ee.conf.Name, ee.server.srv.Config().APUrls)
	}

	// volunteer self and start watching for your nomination
	if err := ee.volunteerSelf(); err != nil {
		ee.Stop()
		return nil, err
	}

	// start campaign to become the leader
	ee.startCampaign()

	return ee, nil
}

// Stop the ElasticEtcd instance
func (ee *ElasticEtcd) Stop() {
	ee.lock.Lock()
	defer ee.lock.Unlock()

	ee.stopping = true
	ee.stopClient()
	ee.stopServer()
	ee.logFile.Close()
}
