package main

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/glustercli/cmd"
)

func main() {
	// Migrate old format Args into new Format. Modifies os.Args[]
	argsMigrate()

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
