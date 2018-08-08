package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/olekukonko/tablewriter"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpPeerCmd       = "Gluster Peer Management"
	helpPeerAddCmd    = "add peer specified by <HOSTNAME>"
	helpPeerRemoveCmd = "remove peer specified by <PeerID>"
	helpPeerStatusCmd = "list status of peers"
	helpPeerListCmd   = "list all the nodes in the pool (including localhost)"
)

var (
	// Peer Remove Command Flags
	flagPeerRemoveForce bool
)

func init() {
	peerCmd.AddCommand(peerAddCmd)

	peerRemoveCmd.Flags().BoolVarP(&flagPeerRemoveForce, "force", "f", false, "Force")

	peerCmd.AddCommand(peerRemoveCmd)

	peerCmd.AddCommand(peerStatusCmd)

	peerListCmd.Flags().StringVar(&flagCmdFilterKey, "key", "", "Filter by metadata key")
	peerListCmd.Flags().StringVar(&flagCmdFilterValue, "value", "", "Filter by metadata value")
	peerCmd.AddCommand(peerListCmd)
}

var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: helpPeerCmd,
}

var peerAddCmd = &cobra.Command{
	Use:   "add <HOSTNAME>",
	Short: helpPeerAddCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostname := cmd.Flags().Args()[0]
		peerAddReq := api.PeerAddReq{
			Addresses: []string{hostname},
		}
		peer, err := client.PeerAdd(peerAddReq)
		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).WithField("host", hostname).Error("peer add failed")
			}
			failure("Peer add failed", err, 1)
		}
		fmt.Println("Peer add successful")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Client Addresses", "Peer Addresses"})
		table.Append([]string{peer.ID.String(), peer.Name, strings.Join(peer.ClientAddresses, "\n"), strings.Join(peer.PeerAddresses, "\n")})
		table.Render()
	},
}

var peerRemoveCmd = &cobra.Command{
	Use:   "remove <PeerID>",
	Short: helpPeerRemoveCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		peerID := cmd.Flags().Args()[0]
		var err error
		if uuid.Parse(peerID) == nil {
			err = errors.New("failed to parse peerID")
		}
		if err == nil {
			err = client.PeerRemove(peerID)
		}
		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).WithField("peerID", peerID).Error("peer remove failed")
			}
			failure("Peer remove failed", err, 1)
		}
		fmt.Println("Peer remove success")
	},
}

func peerStatusHandler(cmd *cobra.Command) {
	var peers api.PeerListResp
	var err error
	if flagCmdFilterKey == "" && flagCmdFilterValue == "" {
		peers, err = client.Peers()
	} else if flagCmdFilterKey != "" && flagCmdFilterValue == "" {
		peers, err = client.Peers(map[string]string{"key": flagCmdFilterKey})
	} else if flagCmdFilterKey == "" && flagCmdFilterValue != "" {
		peers, err = client.Peers(map[string]string{"value": flagCmdFilterValue})
	} else if flagCmdFilterKey != "" && flagCmdFilterValue != "" {
		peers, err = client.Peers(map[string]string{"key": flagCmdFilterKey,
			"value": flagCmdFilterValue,
		})
	}
	if err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).Error("peer status failed")
		}
		failure("Failed to get Peers list", err, 1)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Client Addresses", "Peer Addresses", "Online", "PID"})

	for _, peer := range peers {
		table.Append([]string{peer.ID.String(), peer.Name, strings.Join(peer.ClientAddresses, "\n"), strings.Join(peer.PeerAddresses, "\n"), formatBoolYesNo(peer.Online), formatPID(peer.PID)})
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

var peerListCmd = &cobra.Command{
	Use:   "list",
	Short: helpPeerListCmd,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		peerStatusHandler(cmd)
	},
}
