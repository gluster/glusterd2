package etcdmgmt

import (
	"io/ioutil"
	"net/url"
	"path"

	"github.com/gluster/glusterd2/gdctx"

	"github.com/coreos/etcd/embed"
	"github.com/ghodss/yaml"
	config "github.com/spf13/viper"
)

// EtcdConfigFile is path to etcd config file
var EtcdConfigFile string

// EtcdMinimalConfig represents essential etcd parameters.
type EtcdMinimalConfig struct {
	InitialCluster string `json:"initial-cluster" yaml:"initial-cluster"`
	ClusterState   string `json:"initial-cluster-state" yaml:"initial-cluster-state"`
	Name           string `json:"name" yaml:"name"`
	Dir            string `json:"data-dir" yaml:"data-dir"`
}

// GetNewEtcdConfig will return reference to embed.Config object. This
// is to be passed to embed.StartEtcd() function.
func GetNewEtcdConfig() (*embed.Config, error) {

	// NOTE: This sets most of the fields internally with default values.
	// For example, most of *URL fields are filled with all available IPs
	// of local node i.e binds on all addresses.
	cfg := embed.NewConfig()

	// By convention, human-readable etcd instance names are set to
	// hostname of node. But we need a mapping between peer addresses
	// and their etcd names to make things simple.
	cfg.Name = gdctx.HostIP
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

	return cfg, nil
}

// StoreEtcdConfig stores etcd config info into file
func StoreEtcdConfig(c *embed.Config) error {
	cfg := EtcdMinimalConfig{
		InitialCluster: c.InitialCluster,
		ClusterState:   c.ClusterState,
		Name:           c.Name,
		Dir:            c.Dir,
	}

	y, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(EtcdConfigFile, y, 0644)
	if err != nil {
		return err
	}

	return nil
}

// ReadEtcdConfig reads etcd configuration info stored in file
func ReadEtcdConfig() (*EtcdMinimalConfig, error) {
	y, err := ioutil.ReadFile(EtcdConfigFile)
	if err != nil {
		return nil, err
	}

	var cfg EtcdMinimalConfig

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
