package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpClusterOptionCmd    = "Gluster Cluster Option Management"
	helpClusterOptionSetCmd = "Set Cluster Option"
	helpClusterOptionGetCmd = "Get Cluster Option"
)

func init() {
	clusterCmd.AddCommand(clusterOptionGetCmd)
	clusterCmd.AddCommand(clusterOptionSetCmd)
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: helpClusterOptionCmd,
}

var clusterOptionSetCmd = &cobra.Command{
	Use:   "set <option> <value> [<option> <value>]...",
	Short: helpClusterOptionSetCmd,
	Args:  clusterSetCmdArgs,
	Run:   clusterSetCmdRun,
}

func clusterSetCmdArgs(cmd *cobra.Command, args []string) error {
	// Ensure we have enough arguments for the command
	if len(args) < 2 {
		return errors.New("need at least 2 arguments")
	}

	// Ensure we have a proper option-value pairs
	if (len(args) % 2) != 0 {
		return errors.New("needs '<option> <value>' to be in pairs")
	}

	return nil
}

func clusterSetCmdRun(cmd *cobra.Command, args []string) {
	options := args[:]
	if err := clusterOptionJSONHandler(cmd, options); err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).Error("cluster option set failed")
		}
		failure("Cluster option set failed", err, 1)
	} else {
		fmt.Println("Options set successfully")
	}
}

func clusterOptionJSONHandler(cmd *cobra.Command, options []string) error {
	copt := make(map[string]string)
	for op, val := range options {
		if op%2 == 0 {
			copt[val] = options[op+1]
		}
	}

	err := client.ClusterOptionSet(api.ClusterOptionReq{
		Options: copt})
	return err
}

var clusterOptionGetCmd = &cobra.Command{
	Use:   "get",
	Short: helpClusterOptionGetCmd,
	Run: func(cmd *cobra.Command, args []string) {
		table := tablewriter.NewWriter(os.Stdout)
		opts, err := client.GetClusterOption()
		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).Error("error getting cluster options")
			}
			failure("Error getting cluster options", err, 1)
		}
		table.SetHeader([]string{"Name", "Modified", "Value", "Default Value"})
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		for _, opt := range opts {
			table.Append([]string{opt.Key, formatBoolYesNo(opt.Modified), opt.Value, opt.DefaultValue})
		}
		table.Render()
	},
}
