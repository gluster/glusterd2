package main

import (
	"os"

	"github.com/gluster/glusterd2/glustercli/cmd"
)

func main() {
	filename := os.Args[1]
	cmd.RootCmd.GenBashCompletionFile(filename)
}
