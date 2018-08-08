package cmd

import (
	"github.com/spf13/cobra"
)

const (
	helpSnapshotCmd = "Gluster Snapshot Management"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: helpSnapshotCmd,
}
