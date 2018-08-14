package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	snapshotDeactivateHelpShort = "Deactivate a Gluster Snapshot"
)

var (
	snapshotDeactivateCmd = &cobra.Command{
		Use:   "deactivate <snapname>",
		Short: snapshotDeactivateHelpShort,
		Args:  cobra.ExactArgs(1),
		Run:   snapshotDeactivateCmdRun,
	}
)

func init() {
	snapshotCmd.AddCommand(snapshotDeactivateCmd)
}

func snapshotDeactivateCmdRun(cmd *cobra.Command, args []string) {
	snapname := cmd.Flags().Args()[0]
	err := client.SnapshotDeactivate(snapname)
	if err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).WithField("snapshot", snapname).Error("snapshot deactivation failed")
		}
		failure("snapshot deactivation failed", err, 1)
	}
	fmt.Printf("Snapshot %s deactivated successfully\n", snapname)
}
