package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/gluster/glusterd2/constants"
	"github.com/gluster/glusterd2/pkg/elasticetcd"

	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

const (
	noEmbedOpt       = "noembed"
	etcdEndpointsOpt = "etcdendpoints"
	etcdCURLsOpt     = "etcdcurls"
	etcdPURLsOpt     = "etcdpurls"
	etcdLogFileOpt   = "etcdlogfile"

	defaultEtcdLogFile = "etcd.log"

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
}

// Config is the GD2 store configuration
type Config struct {
	Endpoints []string
	CURLs     []string
	PURLs     []string
	NoEmbed   bool
	UseTLS    bool

	Dir          string
	ConfFile     string
	CertFile     string
	KeyFile      string
	CAFile       string
	ClntCertFile string
	ClntKeyFile  string
}

// NewConfig returns a new store Config with defaults
func NewConfig() *Config {
	return &Config{
		[]string{elasticetcd.DefaultEndpoint},
		[]string{elasticetcd.DefaultCURL},
		[]string{elasticetcd.DefaultPURL},
		false,
		config.GetBool(constants.UseTLS),
		path.Join(config.GetString("localstatedir"), "store"),
		path.Join(config.GetString("localstatedir"), storeConfFile),
		config.GetString(constants.CertFile),
		config.GetString(constants.KeyFile),
		config.GetString(constants.CAFile),
		config.GetString(constants.ClntCertFile),
		config.GetString(constants.ClntKeyFile),
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

	var saveconf bool
	endpoints := config.GetStringSlice(etcdEndpointsOpt)
	if len(endpoints) > 0 {
		saveconf = true
		conf.Endpoints = endpoints
	}

	curls := config.GetStringSlice(etcdCURLsOpt)
	if len(curls) > 0 {
		saveconf = true
		conf.CURLs = curls
	}

	purls := config.GetStringSlice(etcdPURLsOpt)
	if len(purls) > 0 {
		saveconf = true
		conf.PURLs = purls
	}

	certfile := config.GetString(constants.CertFile)
	if len(certfile) > 0 {
		saveconf = true
		conf.CertFile = certfile
	}

	keyfile := config.GetString(constants.KeyFile)
	if len(keyfile) > 0 {
		saveconf = true
		conf.KeyFile = keyfile
	}

	cafile := config.GetString(constants.CAFile)
	if len(cafile) > 0 {
		saveconf = true
		conf.CAFile = cafile
	}

	clntcertfile := config.GetString(constants.ClntCertFile)
	if len(clntcertfile) > 0 {
		saveconf = true
		conf.ClntCertFile = clntcertfile
	}

	clntkeyfile := config.GetString(constants.ClntKeyFile)
	if len(clntkeyfile) > 0 {
		saveconf = true
		conf.ClntKeyFile = clntkeyfile
	}

	if config.IsSet(noEmbedOpt) {
		saveconf = true
		conf.NoEmbed = config.GetBool(noEmbedOpt)
	}
	if config.IsSet(constants.UseTLS) {
		saveconf = true
		conf.UseTLS = config.GetBool(constants.UseTLS)
	}

	if saveconf {
		log.Debug("saving updated store config")
		if err := conf.Save(); err != nil {
			log.WithError(err).Warn("failed to save updated store config")
		}
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
