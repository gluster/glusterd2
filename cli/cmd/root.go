package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd represents main command
var RootCmd = &cobra.Command{
	Use:   "gluster",
	Short: "Gluster Console Manager (command line utility)",
}

var (
	flagXMLOutput  bool
	flagJSONOutput bool
)

func init() {
	initRESTClient()

	// Global flags, applicable for all sub commands
	RootCmd.PersistentFlags().BoolVarP(&flagXMLOutput, "xml", "", false, "XML Output")
	RootCmd.PersistentFlags().BoolVarP(&flagJSONOutput, "json", "", false, "JSON Output")
}

// Execute function parses flags and executes command
func Execute() {
	RootCmd.Execute()
}
