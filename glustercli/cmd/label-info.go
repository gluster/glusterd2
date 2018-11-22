package cmd

import (
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	labelInfoHelpShort = "Get Gluster Label Info"
)

var (
	labelInfoCmd = &cobra.Command{
		Use:   "info <labelname>",
		Short: labelInfoHelpShort,
		Args:  cobra.ExactArgs(1),
		Run:   labelInfoCmdRun,
	}
)

func init() {
	labelCmd.AddCommand(labelInfoCmd)
}

func labelInfoDisplay(info *api.LabelGetResp) {
	fmt.Println()
	fmt.Println("Label Name:", info.Name)
	fmt.Println("Snap Max Hard Limit:", info.SnapMaxHardLimit)
	fmt.Println("Snap Max Soft Limit:", info.SnapMaxSoftLimit)
	fmt.Println("Auto Delete:", info.AutoDelete)
	fmt.Println("Activate On Create:", info.ActivateOnCreate)
	fmt.Println("Snapshot List:", info.SnapList)
	fmt.Println("Description:", info.Description)
	fmt.Println()

	return
}

func labelInfoHandler(cmd *cobra.Command) error {
	var info api.LabelGetResp
	var err error

	labelname := cmd.Flags().Args()[0]
	info, err = client.LabelInfo(labelname)
	if err != nil {
		return err
	}
	labelInfoDisplay(&info)
	return err
}

func labelInfoCmdRun(cmd *cobra.Command, args []string) {
	if err := labelInfoHandler(cmd); err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).Error("error getting label info")
		}
		failure("Error getting Label info", err, 1)
	}
}
