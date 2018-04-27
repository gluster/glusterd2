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
		scheme := "http"
		if flagHTTPS {
			scheme = "https"
		}
		hostname := fmt.Sprintf("%s://%s:%d", scheme, flagHostname, flagPort)
		initRESTClient(hostname, flagCacert, flagInsecure)
	},
}

var (
	flagXMLOutput  bool
	flagJSONOutput bool
	flagHostname   string
	flagHTTPS      bool
	flagPort       int
	flagCacert     string
	flagInsecure   bool
	flagLogLevel   string
	verbose        bool
)

const (
	defaultLogLevel = "INFO"
)

func init() {
	// Global flags, applicable for all sub commands
	RootCmd.PersistentFlags().BoolVarP(&flagXMLOutput, "xml", "", false, "XML Output")
	RootCmd.PersistentFlags().BoolVarP(&flagJSONOutput, "json", "", false, "JSON Output")
	RootCmd.PersistentFlags().StringVarP(&flagHostname, "glusterd-host", "", "localhost", "Glusterd Host")
	RootCmd.PersistentFlags().BoolVarP(&flagHTTPS, "glusterd-https", "", false, "Use HTTPS while connecting to Glusterd")
	RootCmd.PersistentFlags().IntVarP(&flagPort, "glusterd-port", "", 24007, "Glusterd Port")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Log options
	RootCmd.PersistentFlags().StringVarP(&flagLogLevel, logging.LevelFlag, "", defaultLogLevel, logging.LevelHelp)

	// SSL/TLS options
	RootCmd.PersistentFlags().StringVarP(&flagCacert, "cacert", "", "", "Path to CA certificate")
	RootCmd.PersistentFlags().BoolVarP(&flagInsecure, "insecure", "", false,
		"Accepts any certificate presented by the server and any host name in that certificate.")
}
