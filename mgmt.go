// +build !nomgmt

package main

import (
	"github.com/gluster/glusterd2/mgmt"

	flag "github.com/spf13/pflag"
	config "github.com/spf13/viper"
	"github.com/thejerf/suture"
)

func init() {
	flag.Bool("mgmt", false, "Enable libmgmt in GD2")
}

func addMgmtService(s *suture.Supervisor) {
	if config.GetBool("mgmt") {
		s.Add(mgmt.New())
	}
}
