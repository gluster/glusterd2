package main

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/glustercli/cmd"
	"github.com/gluster/glusterd2/pkg/logging"
)

const (
	defaultLogDir   = "./"
	defaultLogFile  = "cli.log"
	defaultLogLevel = "INFO"
)

func main() {
	if err := logging.Init(defaultLogDir, defaultLogFile, defaultLogLevel); err != nil {
		fmt.Println("Error initializing log file ", err)
	}

	// Migrate old format Args into new Format. Modifies os.Args[]
	argsMigrate()

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
