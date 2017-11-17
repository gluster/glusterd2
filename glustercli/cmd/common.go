package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/gluster/glusterd2/pkg/restclient"
)

var client *restclient.Client
var logWriter io.WriteCloser

func initRESTClient(hostname string, cacert string, insecure bool) {
	client = restclient.New(hostname, "", "", cacert, insecure)
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
