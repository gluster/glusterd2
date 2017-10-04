package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
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
	Run: func(cmd *cobra.Command, args []string) {
		validateNArgs(cmd, 1, 1)
		hostname := cmd.Flags().Args()[0]
		peer, err := client.PeerProbe(hostname)
		if err != nil {
			log.WithField("host", hostname).Println("peer probe failed")
			failure(fmt.Sprintf("Peer probe failed with %s", err.Error()), 1)
		}
		fmt.Println("Peer probe success")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Addresses"})
		table.Append([]string{peer.ID.String(), peer.Name, strings.Join(peer.Addresses, ",")})
		table.Render()
	},
}

var peerDetachCmd = &cobra.Command{
	Use:   "detach <HOSTNAME>",
	Short: helpPeerDetachCmd,
	Run: func(cmd *cobra.Command, args []string) {
		validateNArgs(cmd, 1, 1)
		hostname := cmd.Flags().Args()[0]
		err := client.PeerDetach(hostname)
		if err != nil {
			log.WithField("host", hostname).Println("peer detach failed")
			failure(fmt.Sprintf("Peer detach failed with %s", err.Error()), 1)
		}
		fmt.Println("Peer detach success")
	},
}

func peerStatusHandler(cmd *cobra.Command) {
	validateNArgs(cmd, 0, 0)
	peers, err := client.Peers()
	if err != nil {
		log.Println("peer status failed")
		failure(fmt.Sprintf("Error getting Peers list %s", err.Error()), 1)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Addresses"})
	for _, peer := range peers {
		table.Append([]string{peer.ID.String(), peer.Name, strings.Join(peer.Addresses, ",")})
	}
	table.Render()
}

var peerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: helpPeerStatusCmd,
	Run: func(cmd *cobra.Command, args []string) {
		peerStatusHandler(cmd)
	},
}

var poolListCmd = &cobra.Command{
	Use:   "list",
	Short: helpPoolListCmd,
	Run: func(cmd *cobra.Command, args []string) {
		peerStatusHandler(cmd)
	},
}
