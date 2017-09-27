package cmd

import (
	"github.com/spf13/cobra"
	"os"

	"github.com/gluster/glusterd2/pkg/restclient"
)

var client *restclient.Client

func initRESTClient(hostname string) {
	client = restclient.New(hostname, "", "")
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
		if nargs-1 % 2 == 0 {
			cmd.Usage()
			os.Exit(1)
		}
	}
}
