package cmd

import (
	"errors"
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	volumeSetCmdHelpShort = "Set volume options"
	volumeSetCmdHelpLong  = "Set options on a specified gluster volume. Needs a volume name and at least one option-value pair."
)

var (
	flagSetAdv, flagSetExp, flagSetDep bool

	volumeSetCmd = &cobra.Command{
		Use:   "set <volname> <option> <value> [<option> <value>]...",
		Short: volumeSetCmdHelpShort,
		Long:  volumeSetCmdHelpLong,
		Args:  volumeSetCmdArgs,
		Run:   volumeSetCmdRun,
	}
)

func init() {
	volumeSetCmd.Flags().BoolVar(&flagSetAdv, "advanced", false, "Allow setting advanced options")
	volumeSetCmd.Flags().BoolVar(&flagSetExp, "experimental", false, "Allow setting experimental options")
	volumeSetCmd.Flags().BoolVar(&flagSetDep, "deprecated", false, "Allow setting deprecated options")
	volumeCmd.AddCommand(volumeSetCmd)
}

func volumeSetCmdArgs(cmd *cobra.Command, args []string) error {
	// Ensure we have enough arguments for the command
	if len(args) < 3 {
		return errors.New("need at least 3 arguments")
	}

	// Ensure we have a proper option-value pairs
	if (len(args)-1)%2 != 0 {
		return errors.New("needs '<option> <value>' to be in pairs")
	}

	return nil
}

func volumeSetCmdRun(cmd *cobra.Command, args []string) {
	volname := args[0]
	options := args[1:]
	if err := volumeOptionJSONHandler(cmd, volname, options); err != nil {
		if verbose {
			log.WithFields(log.Fields{
				"volume": volname,
				"error":  err.Error(),
			}).Error("volume option set failed")
		}
		failure("Volume option set failed", err, 1)
	} else {
		fmt.Printf("Options set successfully for %s volume\n", volname)
	}
}

func volumeOptionJSONHandler(cmd *cobra.Command, volname string, options []string) error {
	vopt := make(map[string]string)
	for op, val := range options {
		if op%2 == 0 {
			vopt[val] = options[op+1]
		}
	}

	if volname == "all" {
		err := client.GlobalOptionSet(api.GlobalOptionReq{
			Options: vopt,
		})
		return err
	}

	err := client.VolumeSet(volname, api.VolOptionReq{
		Options:      vopt,
		Advanced:     flagSetAdv,
		Experimental: flagSetExp,
		Deprecated:   flagSetDep,
	})
	return err
}
