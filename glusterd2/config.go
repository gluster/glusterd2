package main

import (
	"encoding/json"
	"errors"
	"expvar"
	"net"
	"os"
	"path"
	"path/filepath"

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
	defaultLogLevel = "debug"
	defaultConfName = "glusterd2"
)

// parseFlags sets up the flags and parses them, this needs to be called before any other operation
func parseFlags() {
	flag.String("workdir", "", "Working directory for GlusterD. (default: current directory)")
	flag.String("localstatedir", "", "Directory to store local state information. (default: workdir)")
	flag.String("rundir", "", "Directory to store runtime data.")
	flag.String("config", "", "Configuration file for GlusterD. By default looks for glusterd2.toml in [/usr/local]/etc/glusterd2 and current working directory.")

	flag.String(logging.DirFlag, "", logging.DirHelp+" (default: workdir/log)")
	flag.String(logging.FileFlag, "STDOUT", logging.FileHelp)
	flag.String(logging.LevelFlag, defaultLogLevel, logging.LevelHelp)

	// TODO: Change default to false (disabled) in future.
	flag.Bool("statedump", true, "Enable /statedump endpoint for metrics.")

	flag.String("clientaddress", "", "Address to bind the REST service.")
	flag.String("peeraddress", "", "Address to bind the inter glusterd2 RPC service.")

	// TODO: SSL/TLS is currently only implemented for REST interface
	flag.String("cert-file", "", "Certificate used for SSL/TLS connections from clients to glusterd2.")
	flag.String("key-file", "", "Private key for the SSL/TLS certificate.")

	// PID file
	flag.String("pidfile", "", "PID file path(default: rundir/gluster/glusterd2.pid)")

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

	config.SetDefault("hooksdir", config.GetString("localstatedir")+"/hooks")

	if config.GetString(logging.DirFlag) == "" {
		config.SetDefault(logging.DirFlag, path.Join(config.GetString("localstatedir"), "log"))
	}

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
		port = config.GetString("defaultpeerport")
	}
	config.Set("peeraddress", host+":"+port)

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
	log.WithField("file", config.ConfigFileUsed()).Info("loaded configuration from file")
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
	config.SetConfigName(defaultConfName)

	// Set default config dir and path if default config file exists
	// Chances of not having default config file is only during development
	// Add current directory to this path if default conf file exists
	confdir := defaultConfDir
	conffile := defaultConfDir + "/" + defaultConfName + ".toml"
	if _, err := os.Stat(defaultConfDir + "/" + defaultConfName + ".toml"); os.IsNotExist(err) {
		cdir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return err
		}
		confdir = cdir
		conffile = cdir + "/" + defaultConfName + ".toml"
		log.Info("default config file not found, loading config file from current directory")
	}

	config.AddConfigPath(confdir)
	if err := config.MergeInConfig(); err != nil {
		log.WithError(err).
			WithField("file", conffile).
			Error("failed to read default config file")
		return err
	}

	// If custom configuration is passed
	if confFile != "" {
		config.SetConfigFile(confFile)
		if err := config.MergeInConfig(); err != nil {
			log.WithError(err).
				WithField("file", confFile).
				Error("failed to read config file")
			return err
		}
	}

	// Use config given by flags
	config.BindPFlags(flag.CommandLine)

	// Finally initialize missing config with defaults
	err := setDefaults()

	return err
}
