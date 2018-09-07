package cmd

import (
	"github.com/spf13/cobra"
)

const (
	helpLabelCmd = "Snapshot Label Management"
)

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: helpLabelCmd,
}

func init() {
	snapshotCmd.AddCommand(labelCmd)
}
