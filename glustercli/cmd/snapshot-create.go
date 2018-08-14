package cmd

import (
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	snapshotCreateHelpShort = "Create a Gluster Snapshot"
	snapshotCreateHelpLong  = "Create a Gluster snapshot of the requested volume. By default it creates the snapshot with out timestamp."
)

var (
	flagSnapshotCreateForce       bool
	flagSnapshotCreateTimestamp   bool
	flagSnapshotCreateDescription string

	snapshotCreateCmd = &cobra.Command{
		Use:   "create <snapname> <volname>",
		Short: snapshotCreateHelpShort,
		Long:  snapshotCreateHelpLong,
		Args:  cobra.MinimumNArgs(2),
		Run:   snapshotCreateCmdRun,
	}
)

func init() {
	snapshotCreateCmd.Flags().StringVar(&flagSnapshotCreateDescription, "desctription", "", "Description of snapshot")
	snapshotCreateCmd.Flags().BoolVar(&flagSnapshotCreateForce, "force", false, "Force")
	snapshotCreateCmd.Flags().BoolVar(&flagSnapshotCreateTimestamp, "timestamp", false, "Append timestamp with snap name")

	snapshotCmd.AddCommand(snapshotCreateCmd)
}

func snapshotCreateCmdRun(cmd *cobra.Command, args []string) {
	snapname := args[0]
	volname := args[1]

	req := api.SnapCreateReq{
		VolName:     volname,
		SnapName:    snapname,
		Force:       flagSnapshotCreateForce,
		TimeStamp:   flagSnapshotCreateTimestamp,
		Description: flagSnapshotCreateDescription,
	}

	snap, err := client.SnapshotCreate(req)
	if err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).WithFields(
				log.Fields{
					"volume":   volname,
					"snapshot": snapname,
				}).Error("snapshot creation failed")
		}
		failure("Snapshot creation failed", err, 1)
	}
	vol := snap.VolInfo
	fmt.Printf("%s Snapshot created successfully\n", vol.Name)
	fmt.Println("Snapshot Volume ID: ", vol.ID)
}
