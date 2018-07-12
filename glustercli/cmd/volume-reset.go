package cmd

import (
	"errors"
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	volumeResetCmdHelpShort = "Reset volume options"
	volumeResetCmdHelpLong  = "Reset options on a specified gluster volume. Needs a volume name and at least one option"
)

var (
	flagForce, flagResetAll bool
)

var volumeResetCmd = &cobra.Command{
	Use:   "reset <volname> <options>",
	Short: volumeResetCmdHelpShort,
	Long:  volumeResetCmdHelpLong,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		volname := args[0]
		options := args[1:]
		req := api.VolOptionResetReq{
			Force: flagForce,
			All:   flagResetAll,
		}
		if flagResetAll {
			req.Options = []string{}
		} else {
			if len(args) < 2 {
				failure("Specify atleast one volume option to reset", errors.New("Specify atleast one volume option to reset"), 1)
			} else {
				req.Options = options
			}
		}
		err := client.VolumeReset(volname, req)
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"volume": volname,
					"error":  err.Error(),
				}).Error("volume reset failed")
			}
			failure("Volume reset failed", err, 1)
		}
		fmt.Printf("Volume options reset successfully\n")
	},
}

func init() {
	volumeResetCmd.Flags().BoolVar(&flagForce, "force", false, "Force reset the volume option")
	volumeResetCmd.Flags().BoolVar(&flagResetAll, "all", false, "Reset all the volume options")
	volumeCmd.AddCommand(volumeResetCmd)
}
