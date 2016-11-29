package etcdmgmt

import (
	"errors"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/embed"
)

var etcdInstance = struct {
	sync.Mutex
	etcd *embed.Etcd
}{}

// StartEmbeddedEtcd will start an embedded etcd server using embed.Config
// passed to it. If unsuccessful, this function returns an error.
func StartEmbeddedEtcd(cfg *embed.Config) error {

	etcdInstance.Lock()
	defer etcdInstance.Unlock()

	if etcdInstance.etcd != nil {
		return errors.New("An instance of etcd embedded server is already running")
	}

	// Start embedded etcd server
	etcd, err := embed.StartEtcd(cfg)
	if err != nil {
		return err
	}

	// The returned embed.Etcd.Server instance is not guaranteed to have
	// joined the cluster yet. Wait on the embed.Etcd.Server.ReadyNotify()
	// channel to know when it's ready for use. Stop waiting after an
	// arbitrary timeout (make it configurable?) of 42 seconds.
	select {
	case <-etcd.Server.ReadyNotify():
		log.Info("Etcd embedded server is ready.")
		etcdInstance.etcd = etcd
		return nil
	case <-time.After(42 * time.Second):
		etcd.Server.Stop() // trigger a shutdown
		return errors.New("Etcd embedded server took too long to start")
	case err := <-etcd.Err():
		return err
	}
}

// DestroyEmbeddedEtcd will gracefully shut down the embedded etcd server and
// deletes the etcd data directory.
func DestroyEmbeddedEtcd() error {

	etcdInstance.Lock()
	defer etcdInstance.Unlock()

	if etcdInstance.etcd == nil {
		return errors.New("etcd instance is nil")
	}

	etcdConfig := etcdInstance.etcd.Config()

	etcdInstance.etcd.Close()
	etcdInstance.etcd = nil
	log.Info("Etcd embedded server is stopped.")

	err := os.RemoveAll(etcdConfig.Dir)
	if err != nil {
		return errors.New("Could not delete etcd data dir")
	}

	err = os.RemoveAll(etcdConfig.WalDir)
	if err != nil {
		return errors.New("Could not delete etcd WAL dir")
	}

	os.Remove(EtcdConfigFile)

	return nil
}
