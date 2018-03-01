package main

import (
	"encoding/json"
	"errors"
	"expvar"
	"net"
	"os"
	"path"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/logging"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

var (
	// metrics
	expConfig = expvar.NewMap("config")
)

const (
	defaultLogLevel      = "debug"
	defaultClientAddress = ":24007"
	defaultPeerAddress   = ":24008"

	defaultConfName = "glusterd2"
)

// Slices,Arrays cannot be constants :(
var (
	defaultConfPaths = []string{
		"/etc/glusterd2",
		"/usr/local/etc/glusterd2",
		".",
	}
)

// parseFlags sets up the flags and parses them, this needs to be called before any other operation
func parseFlags() {
	flag.String("workdir", "", "Working directory for GlusterD. (default: current directory)")
	flag.String("localstatedir", "", "Directory to store local state information. (default: workdir)")
	flag.String("rundir", "", "Directory to store runtime data. (default: workdir/run)")
	flag.String("config", "", "Configuration file for GlusterD. By default looks for glusterd2.toml in [/usr/local]/etc/glusterd2 and current working directory.")

	flag.String(logging.DirFlag, "", logging.DirHelp+" (default: workdir/log)")
	flag.String(logging.FileFlag, "STDOUT", logging.FileHelp)
	flag.String(logging.LevelFlag, defaultLogLevel, logging.LevelHelp)

	// TODO: Change default to false (disabled) in future.
	flag.Bool("statedump", true, "Enable /statedump endpoint for metrics.")

	flag.String("clientaddress", defaultClientAddress, "Address to bind the REST service.")
	flag.String("peeraddress", defaultPeerAddress, "Address to bind the inter glusterd2 RPC service.")

	// TODO: SSL/TLS is currently only implemented for REST interface
	flag.String("cert-file", "", "Certificate used for SSL/TLS connections from clients to glusterd2.")
	flag.String("key-file", "", "Private key for the SSL/TLS certificate.")

	store.InitFlags()

	flag.Parse()
}

// setDefaults sets defaults values for config options not available as a flag,
// and flags which don't have default values
func setDefaults() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	wd := config.GetString("workdir")
	if wd == "" {
		config.SetDefault("workdir", cwd)
		wd = cwd
	}

	if config.GetString("localstatedir") == "" {
		config.SetDefault("localstatedir", wd)
	}

	if config.GetString("rundir") == "" {
		config.SetDefault("rundir", path.Join(config.GetString("localstatedir"), "run"))
	}

	if config.GetString(logging.DirFlag) == "" {
		config.SetDefault(logging.DirFlag, path.Join(config.GetString("localstatedir"), "log"))
	}

	// Set default peer port. This shouldn't be configurable.
	config.SetDefault("defaultpeerport", defaultPeerAddress[1:])

	// Set peer address.
	host, port, err := net.SplitHostPort(config.GetString("peeraddress"))
	if err != nil {
		return errors.New("invalid peer address specified")
	}
	if host == "" {
		host = gdctx.HostIP
	}
	if port == "" {
		port = config.GetString("defaultpeerport")
	}
	config.SetDefault("peeraddress", host+":"+port)

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
	l := log.NewEntry(log.StandardLogger())

	for k, v := range config.AllSettings() {
		expConfig.Set(k, valueType{v})
		l = l.WithField(k, v)
	}
	l.Debug("running with configuration")
}

func initConfig(confFile string) error {
	// Read in configuration from file
	// If a config file is not given try to read from default paths
	// If a config file was given, read in configration from that file.
	// If the file is not present panic.

	// Limit config to toml only to avoid confusion with multiple config types
	config.SetConfigType("toml")

	if confFile == "" {
		config.SetConfigName(defaultConfName)
		for _, p := range defaultConfPaths {
			config.AddConfigPath(p)
		}
	} else {
		config.SetConfigFile(confFile)
	}

	if err := config.ReadInConfig(); err != nil {
		if confFile == "" {
			log.WithFields(log.Fields{
				"paths":  defaultConfPaths,
				"config": defaultConfName + ".toml",
				"error":  err,
			}).Debug("failed to read any config files, continuing with defaults")
		} else {
			log.WithError(err).WithField("file", confFile).Error(
				"failed to read given config file")
			return err
		}
	} else {
		log.WithField("file", config.ConfigFileUsed()).Info("loaded configuration from file")
	}

	// Use config given by flags
	config.BindPFlags(flag.CommandLine)

	// Finally initialize missing config with defaults
	if err := setDefaults(); err != nil {
		return err
	}

	dumpConfigToLog()
	return nil
}
