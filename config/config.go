// Package config implements the command line configuration support for GlusterD
//
// Wherever you need to access config, just import config. There is no need to
// instantiate any objects, do parsing or anything else.
package config

import (
	"flag"
)

// All configuration values which should be available for use by other packages need to be defined here as global variables.
var (
	LogLevel    = flag.String("loglevel", "debug", "Log messages upto this level")
	RestAddress = flag.String("restaddress", ":24007", "Address to bind REST endpoint to")
)

func init() {
	flag.Parse()
}
