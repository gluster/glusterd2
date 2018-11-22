package cmd

import (
	"errors"
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	labelResetCmdHelpShort = "Reset value of  label configuratio options"
	labelResetCmdHelpLong  = "Reset options on a specified gluster snapshot label. Needs a label name and at least one option"
)

var labelResetCmd = &cobra.Command{
	Use:   "reset <labelname> <options>",
	Short: labelResetCmdHelpShort,
	Long:  labelResetCmdHelpLong,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var req api.LabelResetReq
		labelname := args[0]
		options := args[1:]
		if len(args) < 2 {
			failure("Specify atleast one label option to reset", errors.New("Specify atleast one label option to reset"), 1)
		} else {
			req.Configurations = options
		}

		err := client.LabelReset(req, labelname)
		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).WithField("label", labelname).Error("label reset failed")
			}
			failure("Snapshot label reset failed", err, 1)
		}
		fmt.Printf("Snapshot label options reset successfully\n")
	},
}

func init() {
	labelCmd.AddCommand(labelResetCmd)
}
