package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpRebalanceCmd       = "Gluster Rebalance"
	helpRebalanceStartCmd  = "Start rebalance operation for gluster volume"
	helpRebalanceStatusCmd = "Status of rebalance operation"
	helpRebalanceStopCmd   = "Stop rebalance operation"
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
				log.WithError(err).WithField("volume", volname).Error("rebalance start failed")
			}
			failure("rebalance start failed", err, 1)
		}
		fmt.Printf("Rebalance for %s started successfully\n", volname)
	},
}

var rebalanceStopCmd = &cobra.Command{
	Use:   "<VOLNAME> stop",
	Short: helpRebalanceStopCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		volname := cmd.Flags().Args()[0]
		err := client.RebalanceStop(volname)
		if err != nil {
			if verbose {
				log.WithError(err).WithField("volume", volname).Error("Rebalance stop failed")
			}
			failure("rebalance start failed", err, 1)
		}
		fmt.Printf("Rebalance for %s stopped successfully\n", volname)
	},
}
