package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	helpPeerCmd       = "Gluster Peer Management"
	helpPeerProbeCmd  = "probe peer specified by <HOSTNAME>"
	helpPeerDetachCmd = "detach peer specified by <HOSTNAME>"
	helpPeerStatusCmd = "list status of peers"
	helpPoolListCmd   = "list all the nodes in the pool (including localhost)"
)

var (
	// Peer Detach Command Flags
	flagPeerDetachForce bool
)

func init() {
	peerCmd.AddCommand(peerProbeCmd)

	peerDetachCmd.Flags().BoolVarP(&flagPeerDetachForce, "force", "f", false, "Force")
	peerCmd.AddCommand(peerDetachCmd)

	peerCmd.AddCommand(peerStatusCmd)

	poolCmd.AddCommand(poolListCmd)

	RootCmd.AddCommand(peerCmd)
	RootCmd.AddCommand(poolCmd)
}

var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: helpPeerCmd,
}

var poolCmd = &cobra.Command{
	Use:   "pool",
	Short: helpPeerCmd,
}

var peerProbeCmd = &cobra.Command{
	Use:   "probe <HOSTNAME>",
	Short: helpPeerProbeCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostname := cmd.Flags().Args()[0]
		peer, err := client.PeerProbe(hostname)
		if err != nil {
			failure("Peer probe failed", err, 1)
		}
		fmt.Println("Peer probe successful")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "PeerAddresses"})
		table.Append([]string{peer.ID.String(), peer.Name, strings.Join(peer.PeerAddresses, ",")})
		table.Render()
	},
}

var peerDetachCmd = &cobra.Command{
	Use:   "detach <HOSTNAME>",
	Short: helpPeerDetachCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostname := cmd.Flags().Args()[0]
		err := client.PeerDetach(hostname)
		if err != nil {
			failure("Peer detach failed", err, 1)
		}
		fmt.Println("Peer detach success")
	},
}

func peerStatusHandler(cmd *cobra.Command) {
	peers, err := client.Peers()
	if err != nil {
		failure("Failed to get Peers list", err, 1)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "PeerAddresses"})
	for _, peer := range peers {
		table.Append([]string{peer.ID.String(), peer.Name, strings.Join(peer.PeerAddresses, ",")})
	}
	table.Render()
}

var peerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: helpPeerStatusCmd,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		peerStatusHandler(cmd)
	},
}

var poolListCmd = &cobra.Command{
	Use:   "list",
	Short: helpPoolListCmd,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		peerStatusHandler(cmd)
	},
}
