// Package config implements the command line configuration support for GlusterD
//
// Wherever you need to access config, just import config. There is no need to
// instantiate any objects, do parsing or anything else.
package config

import (
	"flag"
	"os"
	"path"
)

// All configuration values which should be available for use by other packages need to be defined here as global variables.
// Any configuration that doesn't require custom default values can have their flags initialised right here.
// Any configuration that require custom default values should be initalized manually.
var (
	LogLevel    = flag.String("loglevel", "debug", "Log messages upto this level")
	RestAddress = flag.String("restaddress", ":24007", "Address to bind REST endpoint to")
	RpcAddress  = flag.String("rpcaddress", ":9876", "Address to bind for RPC")

	// Example to start glusterd2 with REST server listening on port 8080:
	// glusterd2 --restaddress=:8080
	// TODO: Parse and remove ':' appropriately

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
	wd, _ := os.Getwd()

	flag.StringVar(&LocalStateDir, "localstatedir", path.Join(wd, "glusterd"), "Directory to store local state information.")
}

func init() {
	initLocalStateDir()

	flag.Parse()
}
