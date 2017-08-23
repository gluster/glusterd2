package cmd

import (
	"github.com/spf13/cobra"
	"os"

	"github.com/gluster/glusterd2/pkg/restclient"
)

var client *restclient.Client

func initRESTClient() {
	client = restclient.New("http://localhost:24007", "", "")
}

func failure(msg string, err int) {
	os.Stderr.WriteString(msg + "\n")
	if err != 0 {
		os.Exit(err)
	}
}

func validateNArgs(cmd *cobra.Command, min int, max int) {
	nargs := len(cmd.Flags().Args())
	if nargs < min || (max != 0 && nargs > max) {
		cmd.Usage()
		os.Exit(1)
	}
}
