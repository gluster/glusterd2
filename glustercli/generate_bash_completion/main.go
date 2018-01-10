package main

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/glustercli/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <filename>\n", os.Args[0])
		os.Exit(1)
	}
	err := cmd.RootCmd.GenBashCompletionFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating bash completion file, error: %v\n", err)
		os.Exit(1)
	}
}
