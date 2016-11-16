package etcdmgmt

import (
	"io/ioutil"
	"path"

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
