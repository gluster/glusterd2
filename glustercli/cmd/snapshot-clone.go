package cmd

import (
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	snapshotCloneHelpShort = "Clone a Gluster Snapshot"
	snapshotCloneHelpLong  = "Clone a Gluster snapshot into a new volume. This new volume will be a writable copy of snapshot."
)

var (
	snapshotCloneCmd = &cobra.Command{
		Use:   "clone <clonename> <snapname>",
		Short: snapshotCloneHelpShort,
		Long:  snapshotCloneHelpLong,
		Args:  cobra.ExactArgs(2),
		Run:   snapshotCloneCmdRun,
	}
)

func init() {
	snapshotCmd.AddCommand(snapshotCloneCmd)
}

func snapshotCloneCmdRun(cmd *cobra.Command, args []string) {
	clonename := args[0]
	snapname := args[1]

	req := api.SnapCloneReq{
		CloneName: clonename,
	}

	vol, err := client.SnapshotClone(snapname, req)
	if err != nil {
		if verbose {
			log.WithFields(log.Fields{
				"clonename": clonename,
				"snapshot":  snapname,
				"error":     err.Error(),
			}).Error("snapshot clone failed")
		}
		failure("Failed to clone Snapshot", err, 1)
	}
	fmt.Printf("New Volume %s cloned from Snapshot %s\n", vol.Name, snapname)
	fmt.Println("Clone Volume ID: ", vol.ID)
}
