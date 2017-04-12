package main

import (
	"net"
	"os"
	"path"

	"github.com/gluster/glusterd2/gdctx"

	log "github.com/Sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

const (
	defaultLogLevel      = "debug"
	defaultClientAddress = ":24007"
	defaultPeerAddress   = ":24008"

	defaultConfName = "glusterd"
)

// Slices,Arrays cannot be constants :(
var (
	defaultConfPaths = []string{
		"/etc/glusterd",
		".",
	}
	defaultEtcdClientAddress = []string{"http://:2379"}
	defaultEtcdPeerAddress   = []string{"http://:2380"}
	defaultEtcdEndpoints     = []string{}
)

// parseFlags sets up the flags and parses them, this needs to be called before any other operation
func parseFlags() {
	flag.String("workdir", "", "Working directory for GlusterD. (default: current directory)")
	flag.String("localstatedir", "", "Directory to store local state information. (default: workdir)")
	flag.String("rundir", "", "Directory to store runtime data. (default: workdir/run)")
	flag.String("logdir", "", "Directory to store logs. (default: workdir/log)")
	flag.String("logfile", "-", "Log file name. (default: STDERR)")
	flag.String("config", "", "Configuration file for GlusterD. By default looks for glusterd.(yaml|toml|json) in /etc/glusterd and current working directory.")
	flag.String("loglevel", defaultLogLevel, "Severity of messages to be logged.")

	flag.String("clientaddress", defaultClientAddress, "Address to bind the REST service.")
	flag.String("peeraddress", defaultPeerAddress, "Address to bind the inter glusterd2 RPC service.")

	// Etcd config options
	flag.StringSlice("etcdclientaddress", defaultEtcdClientAddress, "Address which etcd server will use for peer to peer communication.")
	flag.StringSlice("etcdpeeraddress", defaultEtcdPeerAddress, "Address which etcd server will use to receive etcd client requests.")
	flag.String("etcdlogfile", "etcd.log", "Log file name for logging embedded etcd logs")
	flag.StringSlice("etcdendpoints", defaultEtcdEndpoints, "")

	flag.Parse()
}

// setDefaults sets defaults values for config options not available as a flag,
// and flags which don't have default values
func setDefaults() {
	cwd, err := os.Getwd()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to get current working directory.")
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
		config.SetDefault("rundir", path.Join(wd, "run"))
	}

	if config.GetString("logdir") == "" {
		config.SetDefault("logdir", path.Join(wd, "log"))
	}

	// Set default peer port. This shouldn't be configurable.
	config.SetDefault("defaultpeerport", defaultPeerAddress[1:])

	// Set peer address.
	host, port, err := net.SplitHostPort(config.GetString("peeraddress"))
	if err != nil {
		log.Fatal("Invalid peer address specified.")
	} else {
		if host == "" {
			host = gdctx.HostIP
		}
		if port == "" {
			port = config.GetString("defaultpeerport")
		}
		config.SetDefault("peeraddress", host+":"+port)
	}

	// If no IP is specified for etcd config options (defaults), set those.
	//etcdConfigOptions := []string{"etcdclientaddress", "etcdpeeraddress"}
	//for _, option := range etcdConfigOptions {
	//host, port, err := net.SplitHostPort(config.GetString(option))
	//if err != nil {
	//log.Fatal("Invalid etcd addresses specified.")
	//} else {
	//if host == "" {
	//config.SetDefault(option, gdctx.HostIP+":"+port)
	//}
	//}
	//}
}

func dumpConfigToLog() {
	l := log.NewEntry(log.StandardLogger())

	for k, v := range config.AllSettings() {
		l = l.WithField(k, v)
	}
	l.Debug("running with configuration")
}

func initConfig(confFile string) {
	// Read in configuration from file
	// If a config file is not given try to read from default paths
	// If a config file was given, read in configration from that file.
	// If the file is not present panic.
	if confFile == "" {
		config.SetConfigName(defaultConfName)
		for _, p := range defaultConfPaths {
			config.AddConfigPath(p)
		}
	} else {
		config.SetConfigFile(confFile)
	}

	e := config.ReadInConfig()
	if e != nil {
		if confFile == "" {
			log.WithFields(log.Fields{
				"paths":  defaultConfPaths,
				"config": defaultConfName + ".(toml|yaml|json)",
				"error":  e,
			}).Debug("failed to read any config files, continuing with defaults")
		} else {
			log.WithFields(log.Fields{
				"config": confFile,
				"error":  e,
			}).Fatal("failed to read given config file")
		}
	} else {
		log.WithField("config", config.ConfigFileUsed()).Info("loaded configuration from file")
	}

	// Use config given by flags
	config.BindPFlags(flag.CommandLine)

	// Finally initialize missing config with defaults
	setDefaults()

	dumpConfigToLog()
}
