package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gluster/glusterd2/pkg/logging"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	// variables set by LDFLAGS during build time
	defaultAuthPath = ""
)

var (
	// GlobalFlag have all Global Flags of `glustercli` command set during runtime
	GlobalFlag *GlustercliOption
)

const (
	defaultLogLevel = "INFO"
	defaultTimeout  = 30 // in seconds
)

// NewGlustercliCmd creates the `glustercli` command and its nested children.
func NewGlustercliCmd() *cobra.Command {
	opts := &GlustercliOption{}
	rootCmd := &cobra.Command{
		Use:   "glustercli",
		Short: "Gluster Console Manager (command line utility)",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			opts.Init()
		},
	}
	opts.AddPersistentFlag(rootCmd.PersistentFlags())
	addSubCommands(rootCmd)
	GlobalFlag = opts
	return rootCmd
}

//addSubCommands will add all sub-commands to root glustercli command
func addSubCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(peerCmd)
	rootCmd.AddCommand(bitrotCmd)
	rootCmd.AddCommand(deviceCmd)
	rootCmd.AddCommand(eventsCmd)
	rootCmd.AddCommand(georepCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(volumeCmd)
}

// GlustercliOption will have all global flags set during run time
type GlustercliOption struct {
	XMLOutput  bool
	JSONOutput bool
	Insecure   bool
	Verbose    bool
	Cacert     string
	LogLevel   string
	User       string
	Secret     string
	SecretFile string
	Endpoints  []string
	Timeout    uint
}

//AddPersistentFlag will initialize the Global Flags of root command.
func (gOpt *GlustercliOption) AddPersistentFlag(flagSet *pflag.FlagSet) {
	// Global flags, applicable for all sub commands
	flagSet.BoolVarP(&gOpt.XMLOutput, "xml", "", false, "XML Output")
	flagSet.BoolVarP(&gOpt.JSONOutput, "json", "", false, "JSON Output")
	flagSet.StringSliceVar(&gOpt.Endpoints, "endpoints", []string{"http://127.0.0.1:24007"}, "glusterd2 endpoints")
	flagSet.BoolVarP(&gOpt.Verbose, "verbose", "v", false, "verbose output")
	flagSet.UintVar(&gOpt.Timeout, "timeout", defaultTimeout,
		"overall client timeout (in seconds) which includes time taken to read the response body")

	//user and secret for token authentication
	flagSet.StringVar(&gOpt.User, "user", "glustercli", "Username for authentication")
	flagSet.StringVar(&gOpt.Secret, "secret", "", "Password for authentication")
	flagSet.StringVar(&gOpt.SecretFile, "secret-file", "", "Path to file which contains the secret for authentication")

	// Log options
	flagSet.StringVarP(&gOpt.LogLevel, logging.LevelFlag, "", defaultLogLevel, logging.LevelHelp)

	// SSL/TLS options
	flagSet.StringVarP(&gOpt.Cacert, "cacert", "", "", "Path to CA certificate")
	flagSet.BoolVarP(&gOpt.Insecure, "insecure", "", false,
		"Accepts any certificate presented by the server and any host name in that certificate.")
}

//Init will initialize logging, secret and rest client
func (gOpt *GlustercliOption) Init() {
	//Initialize logging
	if err := logging.Init("", "stdout", gOpt.LogLevel, false); err != nil {
		fmt.Println("Error initializing log file ", err)
	}
	//Initialize Secret
	gOpt.SetSecret()
	//Initializing Rest Client
	initRESTClient(gOpt.Endpoints[0], gOpt.User, gOpt.Secret, gOpt.Cacert, gOpt.Insecure)

}

// SetSecret will Set the secret based on precedence.
// Secret is taken in following order of precedence (highest to lowest):
// --secret
// --secret-file
// GLUSTERD2_AUTH_SECRET (environment variable)
// --secret-file (default path)
//
// NOTE: For simplicity, we don't distinguish between an empty
// value and an unset value.
func (gOpt *GlustercliOption) SetSecret() {
	// --secret
	if gOpt.Secret != "" {
		return
	}

	// --secret-file
	if gOpt.SecretFile != "" {
		data, err := ioutil.ReadFile(gOpt.SecretFile)
		if err != nil {
			failure(fmt.Sprintf("failed to read secret file %s", gOpt.SecretFile), err, 1)
		}
		gOpt.Secret = string(data)
		return
	}

	// GLUSTERD2_AUTH_SECRET
	if secret := os.Getenv("GLUSTERD2_AUTH_SECRET"); secret != "" {
		gOpt.Secret = secret
		return
	}

	// --secret-file (default path)
	data, err := ioutil.ReadFile(defaultAuthPath)
	if err != nil && !os.IsNotExist(err) {
		if gOpt.Verbose {
			log.WithError(err).Error(
				fmt.Sprintf("failed to read default secret file %s", defaultAuthPath))
		}
	}
	gOpt.Secret = string(data)
}
