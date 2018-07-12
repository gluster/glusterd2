package cmd

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpRebalanceCmd       = "Gluster Rebalance"
	helpRebalanceStartCmd  = "Start rebalance session for gluster volume"
	helpRebalanceStatusCmd = "Status of rebalance seesion"
	helpRebalanceStopCmd   = "Stop rebalance session"
)

var (
	flagRebalanceStartCmdForce     bool
	flagRebalanceStartCmdFixLayout bool
)

var rebalanceCmd = &cobra.Command{
	Use:   "volume rebalance",
	Short: helpRebalanceCmd,
}

func init() {

	// Rebalance Start
	rebalanceStartCmd.Flags().BoolVar(&flagRebalanceStartCmdForce, "force", false, "Force")
	rebalanceStartCmd.Flags().BoolVar(&flagRebalanceStartCmdFixLayout, "fixlayout", false, "FixLayout")
	rebalanceCmd.AddCommand(rebalanceStartCmd)

	RootCmd.AddCommand(rebalanceCmd)
}

var rebalanceStartCmd = &cobra.Command{
	Use:   "<VOLNAME> start [flags]",
	Short: helpRebalanceStartCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		volname := cmd.Flags().Args()[0]
		if flagRebalanceStartCmdForce && flagRebalanceStartCmdFixLayout {
			err := errors.New("conflicting options found")
			failure("Please provide only 1 option", err, 1)
		}
		var err error
		if flagRebalanceStartCmdForce {
			err = client.RebalanceStart(volname, "force")
		} else if flagRebalanceStartCmdFixLayout {
			err = client.RebalanceStart(volname, "fix-layout")
		} else {
			err = client.RebalanceStart(volname, "")
		}
		if err != nil {
			if verbose {
				log.WithError(err).WithFields(log.Fields{
					"volume": volname,
				}).Error("rebalance start failed")
			}
			failure("rebalance start failed", err, 1)
		}
		fmt.Printf("Rebalance for %s started successfully\n", volname)
	},
}
