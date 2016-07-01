// Package config implements the command line configuration support for GlusterD
//
// Wherever you need to access config, just import config. There is no need to
// instantiate any objects, do parsing or anything else.
package config

import (
	"flag"
	"strings"
)

// All configuration values which should be available for use by other packages need to be defined here as global variables.
// Any configuration that doesn't require custom default values can have their flags initialised right here.
// Any configuration that require custom default values should be initalized manually.
var (
	LogLevel = flag.String("loglevel", "debug", "Log messages upto this level")

	// A machine can have multiple network interfaces, each with it's own
	// IP address. If IP is not specified, the REST service listens on all
	// available interfaces. If IP is specified, the REST service binds
	// only to that specific interface.
	RestAddress = flag.String("rest-address", ":24007", "IP address of interface and port to bind REST endpoint to")
	RpcAddress  = flag.String("rpc-address", ":9876", "IP address of interface and port to bind RPC service to")
	RpcPort     = strings.Split(*RpcAddress, ":")[1]

	// Example to start glusterd2 with REST server listening on port 8080
	// and only on local ip.
	// glusterd2 -rest-address=127.0.0.1:8080

	/*
		A non-root user can start glusterd2 by setting appropriate
		permissions to the following paths:
		ETCDConfDir: /var/lib/glusterd
		etcdPidDir: /var/run/gluster
		etcdLogDir: /var/log/glusterfs
	*/

	LocalStateDir string
)

func initLocalStateDir() {

	flag.StringVar(&LocalStateDir, "localstatedir", "/var/lib/glusterd", "Directory to store local state information.")
}

func init() {
	initLocalStateDir()

	flag.Parse()
}
