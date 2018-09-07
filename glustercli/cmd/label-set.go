package cmd

import (
	"errors"
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	labelSetHelpShort = "Set a label value"
	labelSetHelpLong  = "Modify one or more label value ."
)

var (
	labelSetCmd = &cobra.Command{
		Use:   "set <labelname> <option> <value> [<option> <value>]...",
		Short: labelSetHelpShort,
		Long:  labelSetHelpLong,
		Args:  labelSetArgsValidate,
		Run:   labelSetCmdRun,
	}
)

func labelSetArgsValidate(cmd *cobra.Command, args []string) error {
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

func init() {
	labelCmd.AddCommand(labelSetCmd)
}

func labelSetCmdRun(cmd *cobra.Command, args []string) {
	labelname := args[0]
	options := args[1:]

	if err := labelSetHandler(cmd, labelname, options); err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).WithField(
				"labelname", labelname).Error("label set failed")
		}
		failure("Label set failed", err, 1)
	} else {
		fmt.Printf("Label Values set successfully for %s label\n", labelname)
	}

}

func labelSetHandler(cmd *cobra.Command, labelname string, options []string) error {
	confs := make(map[string]string)
	for op, val := range options {
		if op%2 == 0 {
			confs[val] = options[op+1]
		}
	}
	req := api.LabelSetReq{
		Configurations: confs,
	}
	err := client.LabelSet(req, labelname)
	return err
}
