package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd represents main command
var RootCmd = &cobra.Command{
	Use:   "gluster",
	Short: "Gluster Console Manager (command line utility)",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initRESTClient(flagHostname)
	},
}

var (
	flagXMLOutput  bool
	flagJSONOutput bool
	flagHostname   string
)

func init() {
	// Global flags, applicable for all sub commands
	RootCmd.PersistentFlags().BoolVarP(&flagXMLOutput, "xml", "", false, "XML Output")
	RootCmd.PersistentFlags().BoolVarP(&flagJSONOutput, "json", "", false, "JSON Output")
	RootCmd.PersistentFlags().StringVarP(&flagHostname, "host", "", "http://localhost:24007", "Host")
}

// Execute function parses flags and executes command
func Execute() {
	RootCmd.Execute()
}
