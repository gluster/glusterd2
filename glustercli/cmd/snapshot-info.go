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

func snapshotInfoCmdRun(cmd *cobra.Command, args []string) {
	err := snapshotInfoHandler(cmd)
	if err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).Error("error getting snapshot info")
		}
		failure("Error getting Snapshot info", err, 1)
	}
}
