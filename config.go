package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
)

const (
	defaultLogLevel    = "debug"
	defaultRestAddress = ":24007"
	defaultRpcAddress  = ":24008"

	defaultConfName = "glusterd"
)

// Slices,Arrays cannot be constants :(
var (
	defaultConfPaths = []string{
		"/etc/glusterd",
		".",
	}
)

// parseFlags sets up the flags and parses them, this needs to be called before any other operation
func parseFlags() {
	cwd, _ := os.Getwd()

	flag.String("localstatedir", cwd, "Directory to store local state information. Defaults to current working directory.")
	flag.String("config", "", "Configuration file for GlusterD. By default looks for glusterd.(yaml|toml|json) in /etc/glusterd and current working directory.")
	flag.String("loglevel", defaultLogLevel, "Severity of messages to be logged.")
	flag.String("restaddress", defaultRestAddress, "Address to bind the REST service.")
	flag.String("rpcaddress", defaultRpcAddress, "Address to bind the RPC service.")

	flag.Parse()
}

// dumpConfig dumps current config to the log
func dumpConfig() {
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

	// Finally use config given by flags
	config.BindPFlags(flag.CommandLine)

	dumpConfig()
}
