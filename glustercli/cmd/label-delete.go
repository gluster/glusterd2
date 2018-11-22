package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	labelDeleteHelpShort = "Delete labels"
)

var (
	labelDeleteCmd = &cobra.Command{
		Use:   "delete <labelname>",
		Short: labelDeleteHelpShort,
		Args:  cobra.ExactArgs(1),
		Run:   labelDeleteCmdRun,
	}
)

func init() {
	labelCmd.AddCommand(labelDeleteCmd)
}

func labelDeleteCmdRun(cmd *cobra.Command, args []string) {
	labelname := args[0]

	if err := client.LabelDelete(labelname); err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).WithField(
				"label", labelname).Error("label delete failed")
		}
		failure("Label delete failed", err, 1)
	}
	fmt.Printf("%s Label deleted successfully\n", labelname)
}
