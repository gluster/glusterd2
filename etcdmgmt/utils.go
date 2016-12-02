package etcdmgmt

import (
	"io/ioutil"
	"net/url"
	"path"

	"github.com/gluster/glusterd2/gdctx"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/embed"
	"github.com/ghodss/yaml"
	config "github.com/spf13/viper"
)

// EtcdConfigFile is path to etcd config file
var EtcdConfigFile string

type etcdMinimalConfig struct {
	InitialCluster string `json:"initial-cluster" yaml:"initial-cluster"`
	ClusterState   string `json:"initial-cluster-state" yaml:"initial-cluster-state"`
	Name           string `json:"name" yaml:"name"`
	Dir            string `json:"data-dir" yaml:"data-dir"`
}

// GetEtcdConfig will return reference to embed.Config object. This
// is to be passed to embed.StartEtcd() function.
func GetEtcdConfig(readConf bool) (*embed.Config, error) {

	// NOTE: This sets most of the fields internally with default values.
	// For example, most of *URL fields are filled with all available IPs
	// of local node i.e binds on all addresses.
	cfg := embed.NewConfig()

	// etcd member names doesn't have to be unique as etcd internally uses
	// UUIDs for member IDs. In practice, etcd instance names are usually
	// set to hostname of node. But we also need to keep the mapping
	// between peers and their etcd names simple. So etcd member names are
	// set to (peer) UUID of glusterd instance.
	cfg.Name = gdctx.MyUUID.String()
	cfg.Dir = cfg.Name + ".etcd"

	listenClientURL, err := url.Parse("http://" + gdctx.HostIP + ":2379")
	if err != nil {
		return nil, err
	}
	cfg.ACUrls = []url.URL{*listenClientURL}
	cfg.LCUrls = []url.URL{*listenClientURL}

	listenPeerURL, err := url.Parse("http://" + gdctx.HostIP + ":2380")
	if err != nil {
		return nil, err
	}
	cfg.APUrls = []url.URL{*listenPeerURL}
	cfg.LPUrls = []url.URL{*listenPeerURL}

	cfg.InitialCluster = cfg.Name + "=" + listenPeerURL.String()
	cfg.ClusterState = embed.ClusterStateFlagNew

	if readConf {
		oldCfg, err := readEtcdConfig()
		if err == nil {
			log.Info("Found saved etcd config file. Using that.")
			cfg.InitialCluster = oldCfg.InitialCluster
			cfg.ClusterState = oldCfg.ClusterState
			cfg.Name = oldCfg.Name
			cfg.Dir = oldCfg.Dir
		}
	}

	return cfg, nil
}

// StoreEtcdConfig stores etcd config info into file
func StoreEtcdConfig(cfg *embed.Config) error {

	emcfg := &etcdMinimalConfig{
		InitialCluster: cfg.InitialCluster,
		ClusterState:   cfg.ClusterState,
		Name:           cfg.Name,
		Dir:            cfg.Dir,
	}

	y, err := yaml.Marshal(emcfg)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(EtcdConfigFile, y, 0644)
	if err != nil {
		return err
	}

	return nil
}

func readEtcdConfig() (*etcdMinimalConfig, error) {
	y, err := ioutil.ReadFile(EtcdConfigFile)
	if err != nil {
		return nil, err
	}

	var cfg etcdMinimalConfig

	err = yaml.Unmarshal(y, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func init() {
	if EtcdConfigFile == "" {
		EtcdConfigFile = path.Join(config.GetString("localstatedir"), "etcd.yaml")
	}
}
