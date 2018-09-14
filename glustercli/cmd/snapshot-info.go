package cmd

import (
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	snapshotInfoHelpShort = "Get Gluster Snapshot Info"
)

var (
	snapshotInfoCmd = &cobra.Command{
		Use:   "info <snapname>",
		Short: snapshotInfoHelpShort,
		Args:  cobra.ExactArgs(1),
		Run:   snapshotInfoCmdRun,
	}
)

func init() {
	snapshotCmd.AddCommand(snapshotInfoCmd)
}

func snapshotInfoDisplay(snap api.SnapGetResp) {
	vol := &snap.VolInfo
	fmt.Println()
	fmt.Println("Snapshot Name:", vol.Name)
	fmt.Println("Snapshot Volume ID:", vol.ID)
	fmt.Println("State:", vol.State)
	fmt.Println("Origin Volume name:", snap.ParentVolName)
	fmt.Println("Snap Creation Time:", snap.CreatedAt.Format("Mon Jan _2 2006 15:04:05 GMT"))
	if vol.Capacity != 0 {
		fmt.Println("Snapshot Volume Capactiy: ", humanReadable(vol.Capacity))
	}
	fmt.Println("Label:", snap.SnapLabel)
	fmt.Println("Snapshot Description:", snap.Description)
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

func snapshotInfoCmdRun(cmd *cobra.Command, args []string) {
	if err := snapshotInfoHandler(cmd); err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).Error("error getting snapshot info")
		}
		failure("Error getting Snapshot info", err, 1)
	}
}
