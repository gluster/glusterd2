package cmd

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpLabelListCmd = "List all Gluster Labels"
)

func init() {

	labelCmd.AddCommand(labelListCmd)

}

func labelListHandler(cmd *cobra.Command) error {
	var infos api.LabelListResp
	var err error
	labelname := cmd.Flags().Args()[0]

	infos, err = client.LabelList(labelname)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	if len(infos) == 0 {
		fmt.Println("There are no labels in the system")
		return nil
	}
	table.SetHeader([]string{"Name"})
	for _, info := range infos {
		table.Append([]string{info.Name})
	}
	table.Render()
	return err
}

var labelListCmd = &cobra.Command{
	Use:   "list",
	Short: helpLabelListCmd,
	Args:  cobra.ExactArgs(1),
	Run:   labelListCmdRun,
}

func labelListCmdRun(cmd *cobra.Command, args []string) {
	if err := labelListHandler(cmd); err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).Error("error getting label list")
		}
		failure("Error getting Label list", err, 1)
	}
}
