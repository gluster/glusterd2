package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	snapshotDeleteHelpShort = "Delete a Gluster Snapshot"
	snapshotDeleteHelpLong  = "Delete a Gluster snapshot of the given name."
)

var (
	snapshotDeleteCmd = &cobra.Command{
		Use:   "delete <snapname>",
		Short: snapshotDeleteHelpShort,
		Long:  snapshotDeleteHelpLong,
		Args:  cobra.ExactArgs(1),
		Run:   snapshotDeleteCmdRun,
	}
)

func init() {
	snapshotCmd.AddCommand(snapshotDeleteCmd)
}

func snapshotDeleteCmdRun(cmd *cobra.Command, args []string) {
	snapname := args[0]

	err := client.SnapshotDelete(snapname)
	if err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).WithField(
				"snapshot", snapname).Error("snapshot delete failed")
		}
		failure("Snapshot delete failed", err, 1)
	}
	fmt.Printf("%s Snapshot deleted successfully\n", snapname)
}
