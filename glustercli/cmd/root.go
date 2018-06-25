package cmd

import (
	"fmt"

	"github.com/gluster/glusterd2/pkg/logging"

	"github.com/spf13/cobra"
)

// RootCmd represents main command
var RootCmd = &cobra.Command{
	Use:   "glustercli",
	Short: "Gluster Console Manager (command line utility)",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := logging.Init("", "stdout", flagLogLevel, false)
		if err != nil {
			fmt.Println("Error initializing log file ", err)
		}
		initRESTClient(flagEndpoints[0], flagUser, flagSecret, flagCacert, flagInsecure)
	},
}

var (
	flagXMLOutput  bool
	flagJSONOutput bool
	flagEndpoints  []string
	flagCacert     string
	flagInsecure   bool
	flagLogLevel   string
	verbose        bool
	flagUser       string
	flagSecret     string
)

const (
	defaultLogLevel = "INFO"
)

func init() {
	// Global flags, applicable for all sub commands
	RootCmd.PersistentFlags().BoolVarP(&flagXMLOutput, "xml", "", false, "XML Output")
	RootCmd.PersistentFlags().BoolVarP(&flagJSONOutput, "json", "", false, "JSON Output")
	RootCmd.PersistentFlags().StringSliceVar(&flagEndpoints, "endpoints", []string{"http://127.0.0.1:24007"}, "glusterd2 endpoints")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	//user and secret for token authentication
	RootCmd.PersistentFlags().StringVar(&flagUser, "user", "glustercli", "Username for authentication")
	RootCmd.PersistentFlags().StringVar(&flagSecret, "secret", "", "Password for authentication")

	// Log options
	RootCmd.PersistentFlags().StringVarP(&flagLogLevel, logging.LevelFlag, "", defaultLogLevel, logging.LevelHelp)

	// SSL/TLS options
	RootCmd.PersistentFlags().StringVarP(&flagCacert, "cacert", "", "", "Path to CA certificate")
	RootCmd.PersistentFlags().BoolVarP(&flagInsecure, "insecure", "", false,
		"Accepts any certificate presented by the server and any host name in that certificate.")
}
