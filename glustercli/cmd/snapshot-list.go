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
		if len(snaps) == 0 {
			fmt.Println("There are no snapshots in the system")
			return nil
		}
		table.SetHeader([]string{"Name", "Origin Volume"})
		for _, snap := range snaps {
			for _, s := range snap.SnapList {
				table.Append([]string{s.VolInfo.Name, snap.ParentName})
			}
		}
	} else {
		if len(snaps) == 0 {
			fmt.Printf("There are no snapshots for volume %s\n", snaps[0].ParentName)
			return nil
		}

		table.SetHeader([]string{"Name"})
		if len(snaps) > 0 {
			for _, entry := range snaps[0].SnapList {
				table.Append([]string{entry.VolInfo.Name})
			}
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
	if err := snapshotListHandler(cmd); err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).Error("error getting snapshot list")
		}
		failure("Error getting Snapshot list", err, 1)
	}
}
