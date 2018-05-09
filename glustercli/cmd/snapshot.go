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
	helpSnapshotCmd           = "Gluster Snapshot Management"
	helpSnapshotCreateCmd     = "Create a Gluster Snapshot"
	helpSnapshotActivateCmd   = "Activate a Gluster Snapshot"
	helpSnapshotDeactivateCmd = "Deactivate a Gluster Snapshot"
	helpSnapshotInfoCmd       = "Get Gluster Snapshot Info"
	helpSnapshotListCmd       = "List all Gluster Snapshots"
)

var (
	// Start Command Flags
	flagSnapshotActivateCmdForce bool
)

func init() {

	snapshotActivateCmd.Flags().BoolVarP(&flagSnapshotActivateCmdForce, "force", "f", false, "Force")
	snapshotCmd.AddCommand(snapshotActivateCmd)
	snapshotCmd.AddCommand(snapshotDeactivateCmd)
	snapshotCmd.AddCommand(snapshotInfoCmd)
	snapshotCmd.AddCommand(snapshotListCmd)

	RootCmd.AddCommand(snapshotCmd)
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: helpSnapshotCmd,
}

var snapshotActivateCmd = &cobra.Command{
	Use:   "activate <snapname> [--force] ",
	Short: helpSnapshotActivateCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

var snapshotDeactivateCmd = &cobra.Command{
	Use:   "deactivate <snapname>",
	Short: helpSnapshotDeactivateCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		snapname := cmd.Flags().Args()[0]
		err := client.SnapshotDeactivate(snapname)
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"snapshot": snapname,
					"error":    err.Error(),
				}).Error("snapshot deactivation failed")
			}
			failure("snapshot deactivation failed", err, 1)
		}
		fmt.Printf("Snapshot %s deactivated successfully\n", snapname)
	},
}

func snapshotInfoDisplay(snap api.SnapGetResp) {
	vol := &snap.VolInfo
	/*
		data := [][]string{
			{vol.Name, vol.Name},
			{"Snapshot Volume ID:", fmt.Sprintln(vol.ID)},
			{"State:", fmt.Sprintln(vol.State)},
			{"Origin Volume name:", snap.ParentVolName},
			{"Snap Creation Time:", "To Be Added"},
			{"Labels:", "To Be Added"},
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoMergeCells(true)
		table.AppendBulk(data)
		table.Render()
		//	table.Append([]string{"Snapshot Volume ID:", string(vol.ID)})
	*/
	fmt.Println()
	fmt.Println("Snapshot Name:", vol.Name)
	fmt.Println("Snapshot Volume ID:", vol.ID)
	fmt.Println("State:", vol.State)
	fmt.Println("Origin Volume name:", snap.ParentVolName)
	fmt.Println("Snap Creation Time:", "To Be Added")
	fmt.Println("Labels:", "To Be Added")
	fmt.Println()

	return
}

func snapshotInfoHandler(cmd *cobra.Command) error {
	var snap api.SnapGetResp
	var err error

	snapname := cmd.Flags().Args()[0]
	snap, err = client.SnapshotInfo(snapname)

	if err != nil {
		return err
	}
	snapshotInfoDisplay(snap)
	return err
}

var snapshotInfoCmd = &cobra.Command{
	Use:   "info <snapname>",
	Short: helpSnapshotInfoCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := snapshotInfoHandler(cmd)
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error getting snapshot info")
			}
			failure("Error getting Snapshot info", err, 1)
		}
	},
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
		table.SetHeader([]string{"Name", "Origin Volume"})
		for _, snap := range snaps {
			for _, entry := range snap.SnapName {
				table.Append([]string{entry, snap.ParentName})
			}
		}
	} else {
		table.SetHeader([]string{"Name"})
		for _, entry := range snaps[0].SnapName {
			table.Append([]string{entry})
		}

	}
	table.Render()
	return err
}

var snapshotListCmd = &cobra.Command{
	Use:   "list [volname]",
	Short: helpSnapshotListCmd,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		err := snapshotListHandler(cmd)
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error getting snapshot list")
			}
			failure("Error getting Snapshot list", err, 1)
		}
	},
}
