package cmd

import (
	"os"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpSnapshotListCmd = "List all Gluster Snapshots"
)

func init() {

	snapshotCmd.AddCommand(snapshotListCmd)

}

func snapshotListHandler(cmd *cobra.Command) error {
	var snaps api.SnapListResp
	var err error
	volname := ""
	if len(cmd.Flags().Args()) > 0 {
		volname = cmd.Flags().Args()[0]
	}

	snaps, err = client.SnapshotList(volname)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	if volname == "" {
		table.SetHeader([]string{"Name", "Origin Volume"})
		for _, snap := range snaps {
			for _, entry := range snap.SnapName {
				table.Append([]string{entry, snap.ParentName})
			}
		}
	} else {
		table.SetHeader([]string{"Name"})
		for _, entry := range snaps[0].SnapName {
			table.Append([]string{entry})
		}

	}
	table.Render()
	return err
}

var snapshotListCmd = &cobra.Command{
	Use:   "list [volname]",
	Short: helpSnapshotListCmd,
	Args:  cobra.RangeArgs(0, 1),
	Run:   snapshotListCmdRun,
}

func snapshotListCmdRun(cmd *cobra.Command, args []string) {
	err := snapshotListHandler(cmd)
	if err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).Error("error getting snapshot list")
		}
		failure("Error getting Snapshot list", err, 1)
	}
}
