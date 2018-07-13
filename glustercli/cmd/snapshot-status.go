package cmd

import (
	"fmt"
	"strconv"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpSnapshotStatusCmd = "Status of all bricks of a Snapshot"
)

func init() {
	snapshotCmd.AddCommand(snapshotStatusCmd)
}

func displaySnapshotStatus(snap api.SnapStatusResp) {

	fmtStrln := "\n%-17s %s   %v"
	fmtStrtb := "\t%-17s %s   %v\n"

	fmt.Printf(fmtStrln, "Snap Name", ":", snap.SnapName)
	fmt.Printf(fmtStrln, "Snap UUID", ":", snap.ID.String())
	fmt.Printf(fmtStrln, "Parent Volume", ":", snap.ParentName)
	fmt.Println()
	for _, entry := range snap.BrickStatus {
		fmt.Printf(fmtStrtb, "Brick Path", ":", entry.Brick.Info.Path)
		fmt.Printf(fmtStrtb, "Host", ":", entry.Brick.Info.Hostname)
		fmt.Printf(fmtStrtb, "online", ":", strconv.FormatBool(entry.Brick.Online))
		fmt.Printf(fmtStrtb, "Pid", ":", entry.Brick.Pid)
		fmt.Printf(fmtStrtb, "Port", ":", entry.Brick.Port)
		fmt.Printf(fmtStrtb, "Device", ":", entry.Brick.Device)
		fmt.Printf(fmtStrtb, "Data Percentage", ":", entry.LvData.DataPercentage)
		fmt.Printf(fmtStrtb, "LV Size", ":", entry.LvData.LvSize)
		fmt.Printf(fmtStrtb, "Pool LV", ":", entry.LvData.PoolLV)
		fmt.Printf(fmtStrtb, "Volume Group", ":", entry.LvData.VgName)
		fmt.Println()
	}
}

func snapshotStatusHandler(cmd *cobra.Command) error {
	var err error
	snapname := ""
	if len(cmd.Flags().Args()) > 0 {
		snapname = cmd.Flags().Args()[0]
	}

	snap, err := client.SnapshotStatus(snapname)
	if err != nil {
		return err
	}

	if snapname == "" {
		// TODO Status for all snapshot
	} else {
		displaySnapshotStatus(snap)
	}
	return err
}

var snapshotStatusCmd = &cobra.Command{
	Use:   "status <snapname>",
	Short: helpSnapshotStatusCmd,
	Args:  cobra.ExactArgs(1),
	Run:   snapshotStatusCmdRun,
}

func snapshotStatusCmdRun(cmd *cobra.Command, args []string) {
	err := snapshotStatusHandler(cmd)
	if err != nil {
		if verbose {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error getting snapshot status")
		}
		failure("Error getting Snapshot status", err, 1)
	}
}
