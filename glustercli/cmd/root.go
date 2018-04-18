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
		err := logging.Init(FlagLogDir, FlagLogFile, FlagLogLevel, false)
		if err != nil {
			fmt.Println("Error initializing log file ", err)
		}
		scheme := "http"
		if FlagHTTPS {
			scheme = "https"
		}
		hostname := fmt.Sprintf("%s://%s:%d", scheme, FlagHostname, FlagPort)
		initRESTClient(hostname, FlagCacert, FlagInsecure)
	},
}

var (
	// FlagXMLOutput to display the output in XML format
	FlagXMLOutput bool
	// FlagJSONOutput to display the output in JSON format
	FlagJSONOutput bool
	// FlagHostname to represents Glusterd Hostname
	FlagHostname string
	// FlagHTTPS represents HTTPs connection to Glusterd
	FlagHTTPS bool
	// FlagPort represents Glusterd port
	FlagPort int
	// FlagCacert represents CA cert path to connect to Glusterd
	FlagCacert string
	// FlagInsecure represents ignore HTTPs certificate
	FlagInsecure bool
	// FlagLogDir represents Log directory
	FlagLogDir string
	// FlagLogFile represents Log file
	FlagLogFile string
	// FlagLogLevel represents Log level
	FlagLogLevel string
)

const (
	defaultLogDir   = "./"
	defaultLogFile  = "glustercli.log"
	defaultLogLevel = "INFO"
)

func init() {
	// Global flags, applicable for all sub commands
	RootCmd.PersistentFlags().BoolVarP(&FlagXMLOutput, "xml", "", false, "XML Output")
	RootCmd.PersistentFlags().BoolVarP(&FlagJSONOutput, "json", "", false, "JSON Output")
	RootCmd.PersistentFlags().StringVarP(&FlagHostname, "glusterd-host", "", "localhost", "Glusterd Host")
	RootCmd.PersistentFlags().BoolVarP(&FlagHTTPS, "glusterd-https", "", false, "Use HTTPS while connecting to Glusterd")
	RootCmd.PersistentFlags().IntVarP(&FlagPort, "glusterd-port", "", 24007, "Glusterd Port")

	// Log options
	RootCmd.PersistentFlags().StringVarP(&FlagLogDir, logging.DirFlag, "", defaultLogDir, logging.DirHelp)
	RootCmd.PersistentFlags().StringVarP(&FlagLogFile, logging.FileFlag, "", defaultLogFile, logging.FileHelp)
	RootCmd.PersistentFlags().StringVarP(&FlagLogLevel, logging.LevelFlag, "", defaultLogLevel, logging.LevelHelp)

	// SSL/TLS options
	RootCmd.PersistentFlags().StringVarP(&FlagCacert, "cacert", "", "", "Path to CA certificate")
	RootCmd.PersistentFlags().BoolVarP(&FlagInsecure, "insecure", "", false,
		"Accepts any certificate presented by the server and any host name in that certificate.")
}
