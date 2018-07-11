package cmd

import (
	fmt

	"github.com/gluster/glusterd2/pkg/restclient"
	rebalance "github.com/gluster/glusterd2/plugins/rebalance/api"
)

const (
	helpRebalanceCmd		= "Gluster Rebalance"
	helpRebalanceStartCmd		= "Start rebalance session for gluster volume"
	helpRebalanceStatusCmd		= "Status of rebalance seesion"
	helpRebalanceStopCmd		= "Stop rebalance session"
)

(
	flagRebalanceStartCmdForce         bool
	flagRebalanceStartCmdFixLayout     bool
)

var rebalanceCmd = &cobra.Command{
	Use:	"volume rebalance",
	Short:	helpRebalanceCmd
}


func init() {

	// Rebalance Start
	rebalanceStartCmd.Flags().BoolVar(&flagRebalanceStartCmdForce, "force", false, "Force")
	rebalanceStartCmd.Flags().BoolVar(&flagRebalanceStartCmdFixLayout, "fixlayout", false, "FixLayout")
	rebalanceCmd.AddCommand(rebalanceStartCmd)

	RootCmd.AddCommand(rebalanceCmd)
}

var rebalaceStartCmd = &cobra.Command{
	Use:	"<VOLNAME> start [flags]",
	Short:	helpRebalanceStartCmd,
	Args:	cobra.ExactArgs(1),
	Run:	func(cmd *cobra.Command, args []string) {
			volname := cmd.Flags().Args()[0]
			err := client.VolumeStart(volname, "")
			if err != nil {
				if verbose {
					log.WithError(err.Error()).WithFields(log.Fields{
						"volume": volname,
					}).Error("rebalance start failed")
			}
			failure("rebalance start failed", err, 1)
		}
		fmt.Printf("Rebalance for %s started successfully\n", volname)
		},
}
