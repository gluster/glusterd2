package elasticetcd

import (
	"errors"
	"io"
	"os"
	"path"
	"time"

	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/pkg/capnslog"
	"github.com/sirupsen/logrus"
)

type server struct {
	srv     *embed.Etcd
	conf    *embed.Config
	logFile io.WriteCloser
}

// startServer starts the embedded etcd server.
// Ensure this is only called with ee.lock held.
func (ee *ElasticEtcd) startServer(initialCluster string) error {
	if ee.server.srv != nil {
		// Do nothing as an etcd server is already running
		return nil
	}

	ee.server.conf = ee.newEmbedConfig(initialCluster)

	ee.log.WithFields(logrus.Fields{
		"name":           ee.server.conf.Name,
		"purls":          types.URLs(ee.server.conf.LPUrls).String(),
		"curls":          types.URLs(ee.server.conf.LCUrls).String(),
		"initialcluster": ee.server.conf.InitialCluster,
		"clusterstate":   ee.server.conf.ClusterState,
		"datadir":        ee.server.conf.Dir,
	}).Debug("prepared embedded etcd config")

	// If starting with non-empty initial cluster, delete the datadir if it
	// exists. The etcdserver will be brought up as a new server and old data
	// being present will prevent it.
	// Starting with an empty initial cluster implies that we are in a single
	// node cluster, so we need to keep the etcd data.
	if initialCluster != "" {
		os.RemoveAll(ee.server.conf.Dir)
	}

	ee.initEtcdLogging()

	etcd, err := embed.StartEtcd(ee.server.conf)
	if err != nil {
		ee.log.WithError(err).Error("failed to start embeeded etcd")
		return err
	}

	// The returned embed.Etcd.Server instance is not guaranteed to have
	// joined the cluster yet. Wait on the embed.Etcd.Server.ReadyNotify()
	// channel to know when it's ready for use. Stop waiting after an
	// arbitrary timeout (make it configurable?) of 42 seconds.
	select {
	case <-etcd.Server.ReadyNotify():
		ee.log.Debug("embedded server ready")
		ee.server.srv = etcd
		return nil
	case <-time.After(42 * time.Second):
		ee.log.Debug("timedout trying to start embedded server")
		etcd.Server.Stop() // trigger a shutdown
		return errors.New("etcd embedded server took too long to start")
	case err := <-etcd.Err():
		return err
	}
}

// stopServer stops the embedded etcd server.
// Ensure this is only called with ee.lock held
func (ee *ElasticEtcd) stopServer() error {
	if ee.server.srv == nil {
		return errors.New("etcd server not running")
	}
	ee.server.srv.Close()
	ee.server.srv = nil
	ee.server.logFile.Close()

	return nil
}

// newEmbedConfig returns a filled embed.Config based on the ElasticEtcd Config and the passed InitialCluster.
// Defaults are filled in as required.
func (ee *ElasticEtcd) newEmbedConfig(initialCluster string) *embed.Config {
	conf := embed.NewConfig()
	conf.Name = ee.conf.Name
	conf.Dir = path.Join(ee.conf.Dir, "etcd.data")

	conf.LCUrls = ee.conf.CURLs
	conf.ACUrls = ee.conf.CURLs
	conf.LPUrls = ee.conf.PURLs
	conf.APUrls = ee.conf.PURLs

	// The default CURL and PURL cannot be used for advertisement, so set the
	// ACurls and APurls to defaultACURLs and defaultAPURLs which are generated
	// from the list of available interfaces
	if isDefaultCURL(conf.ACUrls) {
		conf.ACUrls = defaultACURLs
	}
	if isDefaultPURL(conf.APUrls) {
		conf.APUrls = defaultAPURLs
	}

	if initialCluster != "" {
		conf.InitialCluster = initialCluster
		conf.ClusterState = embed.ClusterStateFlagExisting
	} else {
		ee.log.Debug("initial cluster not given, setting to self")
		conf.InitialCluster = conf.InitialClusterFromName(conf.Name)
	}

	conf.ClientTLSInfo = transport.TLSInfo{
		CertFile:       ee.conf.CertFile,
		KeyFile:        ee.conf.KeyFile,
		CAFile:         ee.conf.CAFile,
		TrustedCAFile:  ee.conf.TrustedCAFile,
		ClientCertAuth: true,
	}
	conf.ClientAutoTLS = true
	conf.PeerTLSInfo = transport.TLSInfo{
		CertFile:       ee.conf.CertFile,
		KeyFile:        ee.conf.KeyFile,
		CAFile:         ee.conf.CAFile,
		TrustedCAFile:  ee.conf.TrustedCAFile,
		ClientCertAuth: true,
	}
	conf.PeerAutoTLS = true

	return conf
}

func (ee *ElasticEtcd) initEtcdLogging() {
	ee.server.logFile = new(nilWriteCloser)
	if !ee.conf.DisableLogging {
		f, err := os.OpenFile(path.Join(ee.conf.LogDir, "etcd.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			ee.server.logFile = f
		}
	}
	capnslog.SetFormatter(capnslog.NewPrettyFormatter(ee.server.logFile, false))
}
