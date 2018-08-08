package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	snapshotRestoreHelpShort = "Restore a Gluster Snapshot to it's parent volume"
	snapshotRestoreHelpLong  = "Parent volume will be restored from the snap, which include data as well as the configuration"
)

var (
	snapshotRestoreCmd = &cobra.Command{
		Use:   "restore <snapname>",
		Short: snapshotRestoreHelpShort,
		Long:  snapshotRestoreHelpLong,
		Args:  cobra.ExactArgs(1),
		Run:   snapshotRestoreCmdRun,
	}
)

func init() {
	snapshotCmd.AddCommand(snapshotRestoreCmd)
}

func snapshotRestoreCmdRun(cmd *cobra.Command, args []string) {
	snapname := cmd.Flags().Args()[0]
	vol, err := client.SnapshotRestore(snapname)
	if err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).WithFields(log.Fields{
				"snapshot": snapname,
				"volume":   vol.Name,
			}).Error("snapshot restore failed")
		}
		failure("snapshot activation failed", err, 1)
	}
	fmt.Printf("Snapshot %s restored successfully to volume %s\n", snapname, vol.Name)
}
