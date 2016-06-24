// Package config implements the command line configuration support for GlusterD
//
// Wherever you need to access config, just import config. There is no need to
// instantiate any objects, do parsing or anything else.
package config

import (
	"flag"
)

// All configuration values which should be available for use by other packages need to be defined here as global variables.
// Any configuration that doesn't require custom default values can have their flags initialised right here.
// Any configuration that require custom default values should be initalized manually.
var (
	LogLevel = flag.String("loglevel", "debug", "Log messages upto this level")

	// A machine can have multiple network interfaces, each with it's own
	// IP address. If -rest-ip is not specified, the REST service listens
	// on all available interfaces. If -rest-ip is specified, the REST
	// service binds only to that specific interface.
	RestIp   = flag.String("rest-ip", "", "IP address of interface to bind REST endpoint to")
	RestPort = flag.String("rest-port", "24007", "Port to bind REST endpoint to")

	RpcIp   = flag.String("rpc-ip", "", "IP address of interface to bind RPC service to")
	RpcPort = flag.String("rpc-port", "9876", "Port to bind for RPC service to")

	// Example to start glusterd2 with REST server listening on port 8080
	// and on local ip.
	// glusterd2 -rest-port=8080 -rest-ip=127.0.0.1

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
