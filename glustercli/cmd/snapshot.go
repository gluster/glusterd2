package cmd

import (
	"github.com/spf13/cobra"
)

const (
	helpSnapshotCmd = "Gluster Snapshot Management"
)

func init() {
	RootCmd.AddCommand(snapshotCmd)
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: helpSnapshotCmd,
}
