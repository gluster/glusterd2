package cmd

import (
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	snapshotActivateHelpShort = "Activate a Gluster Snapshot"
	snapshotActivateHelpLong  = "Activate a Gluster snapshot. Force flag can be used to activate a snapshot forcefully. It will override some checks."
)

var (
	flagSnapshotActivateCmdForce bool

	snapshotActivateCmd = &cobra.Command{
		Use:   "activate <snapname>",
		Short: snapshotActivateHelpShort,
		Long:  snapshotActivateHelpLong,
		Args:  cobra.ExactArgs(1),
		Run:   snapshotActivateCmdRun,
	}
)

func init() {
	snapshotActivateCmd.Flags().BoolVarP(&flagSnapshotActivateCmdForce, "force", "f", false, "Force")
	snapshotCmd.AddCommand(snapshotActivateCmd)
}

func snapshotActivateCmdRun(cmd *cobra.Command, args []string) {
	snapname := cmd.Flags().Args()[0]
	req := api.SnapActivateReq{
		Force: flagSnapshotActivateCmdForce,
	}
	err := client.SnapshotActivate(req, snapname)
	if err != nil {
		if verbose {
			log.WithFields(log.Fields{
				"snapshot": snapname,
				"error":    err.Error(),
			}).Error("snapshot activation failed")
		}
		failure("snapshot activation failed", err, 1)
	}
	fmt.Printf("Snapshot %s activated successfully\n", snapname)
}
