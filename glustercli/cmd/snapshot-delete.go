package cmd

import (
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	snapshotDeleteHelpShort    = "Delete snapshots"
	snapshotDeleteHelpLong     = "It deletes a snapshot if a snapshot name is given as an argument. It also takes command all to deletes all snapshots in a cluster or all snapshots of a particular volume"
	snapshotDeleteAllHelpShort = "Deletes all snapshots"
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

var (
	snapshotDeleteAllCmd = &cobra.Command{
		Use:   "all [volname]",
		Short: snapshotDeleteAllHelpShort,
		Long:  snapshotDeleteAllHelpLong,
		Args:  cobra.MaximumNArgs(1),
		Run:   snapshotDeleteAllCmdRun,
	}
)

func init() {
	snapshotCmd.AddCommand(snapshotDeleteCmd)
	snapshotDeleteCmd.AddCommand(snapshotDeleteAllCmd)
}

func snapshotDelete(snapname string) error {
	if err := client.SnapshotDelete(snapname); err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).WithField(
				"snapshot", snapname).Error("snapshot delete failed")
		}
		return err
	}
	fmt.Printf("%s Snapshot deleted successfully\n", snapname)
	return nil
}

func snapshotDeleteCmdRun(cmd *cobra.Command, args []string) {
	snapname := args[0]
	if err := snapshotDelete(snapname); err != nil {
		failure("Snapshot delete failed", err, 1)
	}
}

func snapshotDeleteAllCmdRun(cmd *cobra.Command, args []string) {
	var snaps api.SnapListResp
	var err error
	volname := ""
	if len(args) > 0 {
		volname = args[0]
	}

	snaps, err = client.SnapshotList(volname)
	if err != nil {
		failure("Failed to get all snapshots", err, 1)
	}

	if len(snaps) == 0 {
		fmt.Printf("There are no snapshots to delete \n")
		return
	}

	for _, vol := range snaps {
		if len(vol.SnapList) == 0 {
			fmt.Printf("There are no snapshots to delete for volume %s \n", vol.ParentName)
			continue
		}
		fmt.Printf("Deleting snapshots of volume %s \n", vol.ParentName)
		for _, snap := range vol.SnapList {
			if err := snapshotDelete(snap.VolInfo.Name); err != nil {
				fmt.Printf("Failed to delete snapshot %s \n", snap.VolInfo.Name)
			}
		}
	}
}
