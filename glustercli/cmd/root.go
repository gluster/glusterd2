package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gluster/glusterd2/pkg/logging"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// variables set by LDFLAGS during build time
	defaultAuthPath = ""
)

var (
	// global variables set during runtime
	secret string
)

const (
	defaultLogLevel = "INFO"
	defaultTimeout  = 30 // in seconds
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

		// Secret is taken in following order of precedence (highest to lowest):
		// --secret
		// --secret-file
		// GLUSTERD2_AUTH_SECRET (environment variable)
		// --secret-file (default path)

		// NOTE: For simplicity, we don't distinguish between an empty
		// value and an unset value.

		// --secret
		if flagSecret != "" {
			secret = flagSecret
		}

		// --secret-file
		if flagSecretFile != "" && secret == "" {
			data, err := ioutil.ReadFile(flagSecretFile)
			if err != nil {
				failure(fmt.Sprintf("failed to read secret file %s", flagSecretFile),
					err, 1)
			}
			secret = string(data)
		}

		// GLUSTERD2_AUTH_SECRET
		if secret == "" {
			secret = os.Getenv("GLUSTERD2_AUTH_SECRET")
		}

		// --secret-file (default path)
		if flagSecretFile == "" && secret == "" {
			data, err := ioutil.ReadFile(defaultAuthPath)
			if err != nil && !os.IsNotExist(err) {
				if verbose {
					log.WithError(err).Error(
						fmt.Sprintf("failed to read default secret file %s", defaultAuthPath))
				}
			}
			secret = string(data)
		}

		initRESTClient(flagEndpoints[0], flagUser, secret, flagCacert, flagInsecure)
	},
}

var (
	// set by command line flags
	flagXMLOutput  bool
	flagJSONOutput bool
	flagInsecure   bool
	verbose        bool
	flagCacert     string
	flagLogLevel   string
	flagUser       string
	flagSecret     string
	flagSecretFile string
	flagEndpoints  []string
	flagTimeout    uint
)

func init() {
	// Global flags, applicable for all sub commands
	RootCmd.PersistentFlags().BoolVarP(&flagXMLOutput, "xml", "", false, "XML Output")
	RootCmd.PersistentFlags().BoolVarP(&flagJSONOutput, "json", "", false, "JSON Output")
	RootCmd.PersistentFlags().StringSliceVar(&flagEndpoints, "endpoints", []string{"http://127.0.0.1:24007"}, "glusterd2 endpoints")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().UintVar(&flagTimeout, "timeout", defaultTimeout,
		"overall client timeout (in seconds) which includes time taken to read the response body")

	//user and secret for token authentication
	RootCmd.PersistentFlags().StringVar(&flagUser, "user", "glustercli", "Username for authentication")
	RootCmd.PersistentFlags().StringVar(&flagSecret, "secret", "", "Password for authentication")
	RootCmd.PersistentFlags().StringVar(&flagSecretFile, "secret-file", "", "Path to file which contains the secret for authentication")

	// Log options
	RootCmd.PersistentFlags().StringVarP(&flagLogLevel, logging.LevelFlag, "", defaultLogLevel, logging.LevelHelp)

	// SSL/TLS options
	RootCmd.PersistentFlags().StringVarP(&flagCacert, "cacert", "", "", "Path to CA certificate")
	RootCmd.PersistentFlags().BoolVarP(&flagInsecure, "insecure", "", false,
		"Accepts any certificate presented by the server and any host name in that certificate.")
}
