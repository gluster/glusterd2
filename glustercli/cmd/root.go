package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gluster/glusterd2/pkg/logging"

	log "github.com/sirupsen/logrus"
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

		if flagAuthFile == "" && flagSecret == "" {
			data, err := ioutil.ReadFile(defaultAuthPath)
			if err != nil && !os.IsNotExist(err) {
				if verbose {
					log.WithError(err).Error("failed to read secret")
				}
			}
			secret = string(data)
		}

		if flagAuthFile != "" {
			data, err := ioutil.ReadFile(flagAuthFile)
			if err != nil {
				if verbose {
					log.WithError(err).Error("failed to read secret")
				}
				failure("failed to read secret", err, 1)
			}
			secret = string(data)
		}

		if flagSecret != "" {
			secret = flagSecret
		}

		initRESTClient(flagEndpoints[0], flagUser, secret, flagCacert, flagInsecure)
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
	flagAuthFile   string
	secret         string
	//defaultAuthPath is set by LDFLAGS
	defaultAuthPath = ""
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
	RootCmd.PersistentFlags().StringVar(&flagAuthFile, "authfile", "", "Auth file path, which contains secret for authentication")

	// Log options
	RootCmd.PersistentFlags().StringVarP(&flagLogLevel, logging.LevelFlag, "", defaultLogLevel, logging.LevelHelp)

	// SSL/TLS options
	RootCmd.PersistentFlags().StringVarP(&flagCacert, "cacert", "", "", "Path to CA certificate")
	RootCmd.PersistentFlags().BoolVarP(&flagInsecure, "insecure", "", false,
		"Accepts any certificate presented by the server and any host name in that certificate.")
}
