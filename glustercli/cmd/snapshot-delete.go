package cmd

import (
	"fmt"

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
	flagSnapshotDeleteAllVolume string
	flagSnapshotDeleteAllLabel  string
	snapshotDeleteAllCmd        = &cobra.Command{
		Use:   "all",
		Short: snapshotDeleteAllHelpShort,
		Long:  snapshotDeleteAllHelpLong,
		Args:  cobra.MaximumNArgs(0),
		Run:   snapshotDeleteAllCmdRun,
	}
)

func init() {
	snapshotCmd.AddCommand(snapshotDeleteCmd)
	snapshotDeleteCmd.AddCommand(snapshotDeleteAllCmd)
	snapshotDeleteAllCmd.Flags().StringVar(&flagSnapshotDeleteAllVolume, "volume", "", "Deletes all snapshots of a given volume")
	snapshotDeleteAllCmd.Flags().StringVar(&flagSnapshotDeleteAllLabel, "label", "", "Deletes all snapshots attached to the label")
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

func deleteAllSnapshots(snapList []string) {
	for _, snap := range snapList {
		if err := snapshotDelete(snap); err != nil {
			fmt.Printf("Failed to delete snapshot %s \n", snap)
		}
	}
}

func snapshotDeleteAllCmdRun(cmd *cobra.Command, args []string) {
	var err error
	volname := flagSnapshotDeleteAllVolume
	labelname := flagSnapshotDeleteAllLabel

	if labelname != "" {
		info, err := client.LabelInfo(labelname)
		if err != nil {
			failure("Failed to get all snapshots", err, 1)
		}
		fmt.Printf("Deleting All snapshots of label %s \n", labelname)
		deleteAllSnapshots(info.SnapList)
		if volname != "" {
			//If volume falg has not given then we are done with
			//delete all.
			return
		}
	}

	snaps, err := client.SnapshotList(volname)
	if err != nil {
		failure("Failed to get all snapshots", err, 1)
	}

	if len(snaps) == 0 {
		fmt.Printf("There are no more snapshots to delete \n")
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
