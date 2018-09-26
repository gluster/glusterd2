package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	snapshotDeleteHelpShort    = "Delete snapshots"
	snapshotDeleteHelpLong     = "It deletes a snapshot if a snapshot name is given as an argument. It also takes command all to deletes all snapshots in a cluster or all snapshots of a particular volume"
	snapshotDeleteAllHelpShort = "Deletes All snapshots"
	snapshotDeleteAllHelpLong  = "Deletes all snapshots in the cluster. If volume name is not given. If the volume name is given, then it deletes all snapshots of a volume"
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

	if err := client.SnapshotDelete(snapname); err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).WithField(
				"snapshot", snapname).Error("snapshot delete failed")
		}
		failure("Snapshot delete failed", err, 1)
	}
	fmt.Printf("%s Snapshot deleted successfully\n", snapname)
}
