package cmd

import (
	"fmt"

	"github.com/gluster/glusterd2/pkg/logging"

	"github.com/spf13/cobra"
)

// RootCmd represents main command
var RootCmd = &cobra.Command{
	Use:   "gluster",
	Short: "Gluster Console Manager (command line utility)",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := logging.Init(flagLogDir, flagLogFile, flagLogLevel)
		if err != nil {
			fmt.Println("Error initializing log file ", err)
		}
		initRESTClient(flagHostname, flagCacert, flagInsecure)
	},
}

var (
	flagXMLOutput  bool
	flagJSONOutput bool
	flagHostname   string
	flagCacert     string
	flagInsecure   bool
	flagLogDir     string
	flagLogFile    string
	flagLogLevel   string
)

const (
	defaultLogDir   = "./"
	defaultLogFile  = "glustercli.log"
	defaultLogLevel = "INFO"
)

func init() {
	// Global flags, applicable for all sub commands
	RootCmd.PersistentFlags().BoolVarP(&flagXMLOutput, "xml", "", false, "XML Output")
	RootCmd.PersistentFlags().BoolVarP(&flagJSONOutput, "json", "", false, "JSON Output")
	RootCmd.PersistentFlags().StringVarP(&flagHostname, "host", "", "http://localhost:24007", "Host")

	// Log options
	RootCmd.PersistentFlags().StringVarP(&flagLogDir, logging.DirFlag, "", defaultLogDir, logging.DirHelp)
	RootCmd.PersistentFlags().StringVarP(&flagLogFile, logging.FileFlag, "", defaultLogFile, logging.FileHelp)
	RootCmd.PersistentFlags().StringVarP(&flagLogLevel, logging.LevelFlag, "", defaultLogLevel, logging.LevelHelp)

	// SSL/TLS options
	RootCmd.PersistentFlags().StringVarP(&flagCacert, "cacert", "", "", "Path to CA certificate")
	RootCmd.PersistentFlags().BoolVarP(&flagInsecure, "insecure", "", false,
		"Accepts any certificate presented by the server and any host name in that certificate.")
}

// Execute function parses flags and executes command
func Execute() {
	RootCmd.Execute()
}
