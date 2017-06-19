package elasticetcd

import (
	"errors"
	"io"
	"os"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/pkg/capnslog"
)

var (
	ErrClientNotAvailable = errors.New("etcd client not available")
	ErrAddingToServerList = errors.New("failed to add self to server list")
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

	// Delete the datadir if it exists. The etcdserver will be brought up as a
	// new server and old data being present will prevent it.
	os.RemoveAll(ee.server.conf.Dir)

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
		return errors.New("Etcd embedded server took too long to start")
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

func (ee *ElasticEtcd) newEmbedConfig(initialCluster string) *embed.Config {
	conf := embed.NewConfig()
	conf.Name = ee.conf.Name
	conf.Dir = path.Join(ee.conf.Dir, "etcd.data")

	conf.LCUrls = ee.conf.CURLs
	if isDefaultCURL(ee.conf.CURLs) {
		conf.ACUrls = defaultACURLs
	} else {
		conf.ACUrls = ee.conf.CURLs
	}

	conf.LPUrls = ee.conf.PURLs
	if isDefaultPURL(ee.conf.PURLs) {
		conf.APUrls = defaultAPURLs
	} else {
		conf.APUrls = ee.conf.PURLs
	}

	if initialCluster != "" {
		conf.InitialCluster = initialCluster
		conf.ClusterState = embed.ClusterStateFlagExisting
	} else {
		ee.log.Debug("initial cluster not given, setting to self")
		conf.InitialCluster = conf.InitialClusterFromName(conf.Name)
	}
	return conf
}

func (ee *ElasticEtcd) initEtcdLogging() {
	ee.server.logFile = new(nilWriteCloser)
	if !ee.conf.DisableLogging {
		f, err := os.OpenFile(path.Join(ee.conf.Dir, "etcd.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			ee.server.logFile = f
		}
	}
	capnslog.SetFormatter(capnslog.NewPrettyFormatter(ee.server.logFile, false))
}
