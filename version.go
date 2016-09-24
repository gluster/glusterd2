package main

import (
	"fmt"

	"github.com/gluster/glusterd2/gdctx"

	flag "github.com/spf13/pflag"
)

func init() {
	//Register the `--version` flag
	flag.Bool("version", false, "Show the version information")
}

func dumpVersionInfo() {
	fmt.Println(gdctx.GlusterdVersion)
}
