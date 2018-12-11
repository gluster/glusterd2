package main

import (
	"encoding/json"
	"errors"
	"expvar"
	"net"
	"path"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/logging"
	"github.com/gluster/glusterd2/pkg/tracing"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

const (
	defaultlogfile       = "STDOUT"
	defaultpeerport      = "24008"
	defaultpeeraddress   = ":24008"
	defaultclientaddress = ":24007"
	defaultloglevel      = "debug"
)

var (
	// metrics
	expConfig = expvar.NewMap("config")

	// defaultPathPrefix is set by LDFLAGS
	defaultPathPrefix = ""

	defaultlocalstatedir = path.Join(defaultPathPrefix, "/var/lib/glusterd2")
	defaultlogdir        = path.Join(defaultPathPrefix, "/var/log/glusterd2")
	defaultrundir        = path.Join(defaultPathPrefix, "/var/run/glusterd2")
)

// initFlags sets up the flags and parses them, this needs to be called before any other operation
func initFlags() {
	flag.String("localstatedir", defaultlocalstatedir, "Directory to store local state information.")
	flag.String("rundir", defaultrundir, "Directory to store runtime data.")
	flag.String("config", "", "Configuration file for GlusterD.")

	flag.String(logging.DirFlag, defaultlogdir, logging.DirHelp)
	flag.String(logging.FileFlag, defaultlogfile, logging.FileHelp)
	flag.String(logging.LevelFlag, defaultloglevel, logging.LevelHelp)

	// TODO: Change default to false (disabled) in future.
	flag.Bool("statedump", true, "Enable /statedump endpoint for metrics.")

	flag.String("clientaddress", defaultclientaddress, "Address to bind the REST service.")
	flag.String("peeraddress", defaultpeeraddress, "Address to bind the inter glusterd2 RPC service.")

	// TODO: SSL/TLS is currently only implemented for REST interface
	flag.String("cert-file", "", "Certificate used for SSL/TLS connections from clients to glusterd2.")
	flag.String("key-file", "", "Private key for the SSL/TLS certificate.")

	// PID file
	flag.String("pidfile", "", "PID file path. (default \"rundir/glusterd2.pid)\"")

	store.InitFlags()
	tracing.InitFlags()

	flag.Parse()
}

// setDefaults sets defaults values for config options not available as a flag,
// and flags which don't have default values
func setDefaults() error {

	config.SetDefault("hooksdir", config.GetString("localstatedir")+"/hooks")

	if config.GetString("pidfile") == "" {
		config.SetDefault("pidfile", path.Join(config.GetString("rundir"), "glusterd2.pid"))
	}

	// Set peer address.
	host, port, err := net.SplitHostPort(config.GetString("peeraddress"))
	if err != nil {
		return errors.New("invalid peer address specified")
	}
	if host == "" {
		host = gdctx.HostIP
	}
	if port == "" {
		port = defaultpeerport
	}

	config.Set("peeraddress", host+":"+port)
	config.Set("defaultpeerport", defaultpeerport)

	return nil
}

type valueType struct {
	v interface{}
}

func (v valueType) String() string {
	vb, _ := json.Marshal(v.v)
	return string(vb)
}

func dumpConfigToLog() {
	if config.ConfigFileUsed() != "" {
		log.WithField("file", config.ConfigFileUsed()).Info("loaded configuration from file")
	}

	l := log.NewEntry(log.StandardLogger())

	for k, v := range config.AllSettings() {
		expConfig.Set(k, valueType{v})
		l = l.WithField(k, v)
	}
	l.Debug("running with configuration")
}

// initConfig intializes GD2 configuration from various sources.
// The order of preference is,
// - explicitly set configs using config.Set
// - flags, if set
// - environment variables
// - config file
// - defaults set using config.SetDefault
// - flag defaults
func initConfig() error {
	// Use config given by flags
	config.BindPFlags(flag.CommandLine)

	// Allow config values from environment environment variables.
	// All options settable from the command line are available to be set this way.
	// The environment variable should be in uppercase, prefixed with "GD2" and have "-" replaced by "_" to be used.
	config.SetEnvPrefix("GD2")
	config.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	config.AutomaticEnv()

	// Read in configuration from file
	// If a config file is not given try to read from default paths
	// If a config file was given, read in configration from that file.
	// If the file is not present panic.

	// Limit config to toml only to avoid confusion with multiple config types
	config.SetConfigType("toml")

	confFile := config.GetString("config")
	// If custom configuration is passed use it, if not try to use defaults
	if confFile != "" {
		config.SetConfigFile(confFile)
	} else {
		config.AddConfigPath(path.Join(defaultPathPrefix, "/etc/glusterd2"))
		config.SetConfigName("glusterd2")
	}
	if err := config.ReadInConfig(); err != nil {
		// Ignore error if config file is not found, error out otherwise
		if _, ok := err.(config.ConfigFileNotFoundError); ok {
			log.WithError(err).
				WithField("file", config.ConfigFileUsed()).
				Warn("failed to load config from file")
		} else {
			log.WithError(err).
				WithField("file", config.ConfigFileUsed()).
				Error("failed to load config from file")
			return err
		}
	}

	// Finally initialize missing config with defaults
	err := setDefaults()

	return err
}
