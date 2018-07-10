package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/gluster/glusterd2/pkg/elasticetcd"

	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

const (
	// etcd client options
	noEmbedOpt = "noembed"
	// TODO: Remove noembed as config option, presence of "etcdendpoints" would be sufficient
	etcdEndpointsOpt      = "etcdendpoints"
	etcdClientCertFileOpt = "etcd-client-cert-file"
	etcdClientKeyFileOpt  = "etcd-client-key-file"
	etcdClientCAFileOpt   = "etcd-client-ca-file"

	// etcd server (elasticetcd) options
	etcdCURLsOpt       = "etcdcurls"
	etcdPURLsOpt       = "etcdpurls"
	etcdLogFileOpt     = "etcdlogfile"
	defaultEtcdLogFile = "etcd.log"

	// TODO: Fix these too. Make elasticetcd support TLS if it doesn't
	// already.
	useTLSOpt   = "usetls"
	caFileOpt   = "ca-file"
	certFileOpt = "cert-file"
	keyFileOpt  = "key-file"

	// common options
	storeConfFile = "store.toml"
)

// InitFlags intializes the command line options for the GD2 store
func InitFlags() {
	flag.Bool(noEmbedOpt, false, "Disable the embedded etcd. If disabled --etcdendpoints must be provided.")
	// Not setting defaults for the options here as the defaults will be returned
	// by `config` when nothing has been set overwriting anything saved
	flag.StringSlice(etcdEndpointsOpt, nil, fmt.Sprintf("ETCD endpoints of a remote etcd cluster for the store to connect to. (Defaults to: %s)", elasticetcd.DefaultEndpoint))
	flag.StringSlice(etcdCURLsOpt, nil, fmt.Sprintf("URLs which etcd server will use for peer to peer communication. (Defaults to: %s)", elasticetcd.DefaultCURL))
	flag.StringSlice(etcdPURLsOpt, nil, fmt.Sprintf("URLs which etcd server will use to receive etcd client requests. (Defaults to: %s)", elasticetcd.DefaultPURL))

	flag.String(etcdClientCertFileOpt, "", "identify secure etcd client using this TLS certificate file")
	flag.String(etcdClientKeyFileOpt, "", "identify secure etcd client using this TLS key file")
	flag.String(etcdClientCAFileOpt, "", "verify certificates of TLS-enabled secure etcd servers using this CA bundle")
}

// Config is the GD2 store configuration
type Config struct {
	Endpoints []string
	CURLs     []string
	PURLs     []string
	NoEmbed   bool
	UseTLS    bool
	Dir       string
	ConfFile  string

	// etcd server configuration
	CertFile string
	KeyFile  string
	CAFile   string

	// etcd client configuration
	ClntCertFile string
	ClntKeyFile  string
	ClntCAFile   string
}

// TODO: This is also a mess. We should just create a package level global
// instance of *Config and pass its fields directly to flag.* functions.

// NewConfig returns a new store Config with defaults
func NewConfig() *Config {
	return &Config{
		Endpoints:    []string{elasticetcd.DefaultEndpoint},
		CURLs:        []string{elasticetcd.DefaultCURL},
		PURLs:        []string{elasticetcd.DefaultPURL},
		NoEmbed:      false,
		UseTLS:       false,
		Dir:          path.Join(config.GetString("localstatedir"), "store"),
		ConfFile:     path.Join(config.GetString("localstatedir"), storeConfFile),
		CertFile:     config.GetString(certFileOpt),
		KeyFile:      config.GetString(keyFileOpt),
		CAFile:       config.GetString(caFileOpt),
		ClntCertFile: config.GetString(etcdClientCertFileOpt),
		ClntKeyFile:  config.GetString(etcdClientKeyFileOpt),
		ClntCAFile:   config.GetString(etcdClientCAFileOpt),
	}
}

// Save saves the store config to a file in the localstatedir
func (c *Config) Save() error {
	b, err := toml.Marshal(*c)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(c.ConfFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return nil
}

// GetConfig returns a filled store config
// The config is filled with values from the following sources in order of preference,
// 	- GD2 config
// 	- Store config file
// 	- Defaults
func GetConfig() *Config {
	conf, err := readConfigFile()
	if err != nil {
		log.WithError(err).Warn("could not read store config file, continuing with defaults")
		conf = NewConfig()
	}

	endpoints := config.GetStringSlice(etcdEndpointsOpt)
	if len(endpoints) > 0 {
		conf.Endpoints = endpoints
	}

	curls := config.GetStringSlice(etcdCURLsOpt)
	if len(curls) > 0 {
		conf.CURLs = curls
	}

	purls := config.GetStringSlice(etcdPURLsOpt)
	if len(purls) > 0 {
		conf.PURLs = purls
	}

	certfile := config.GetString(certFileOpt)
	if len(certfile) > 0 {
		conf.CertFile = certfile
	}

	keyfile := config.GetString(keyFileOpt)
	if len(keyfile) > 0 {
		conf.KeyFile = keyfile
	}

	cafile := config.GetString(etcdClientCAFileOpt)
	if len(cafile) > 0 {
		conf.ClntCAFile = cafile
	}

	clntcertfile := config.GetString(etcdClientCertFileOpt)
	if len(clntcertfile) > 0 {
		conf.ClntCertFile = clntcertfile
	}

	clntkeyfile := config.GetString(etcdClientKeyFileOpt)
	if len(clntkeyfile) > 0 {
		conf.ClntKeyFile = clntkeyfile
	}

	if config.IsSet(noEmbedOpt) {
		conf.NoEmbed = config.GetBool(noEmbedOpt)
	}
	if config.IsSet(useTLSOpt) {
		conf.UseTLS = config.GetBool(useTLSOpt)
	}

	log.Debug("saving updated store config")
	if err := conf.Save(); err != nil {
		log.WithError(err).Warn("failed to save updated store config")
	}

	return conf
}

func readConfigFile() (*Config, error) {
	storeConfPath := path.Join(config.GetString("localstatedir"), storeConfFile)

	b, err := ioutil.ReadFile(storeConfPath)
	if err != nil {
		return nil, err
	}

	conf := &Config{}

	if err := toml.Unmarshal(b, conf); err != nil {
		return nil, err
	}

	return conf, nil
}
