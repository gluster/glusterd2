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
	helpPeerRemoveCmd = "remove peer specified by <HOSTNAME or PeerID>"
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

	RootCmd.AddCommand(peerCmd)
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
			if verbose {
				log.WithFields(log.Fields{
					"host":  hostname,
					"error": err.Error(),
				}).Error("peer add failed")
			}
			failure("Peer add failed", err, 1)
		}
		fmt.Println("Peer add successful")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Peer Addresses"})
		table.Append([]string{peer.ID.String(), peer.Name, strings.Join(peer.PeerAddresses, ",")})
		table.Render()
	},
}

var peerRemoveCmd = &cobra.Command{
	Use:   "remove <HOSTNAME or PeerID>",
	Short: helpPeerRemoveCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostname := cmd.Flags().Args()[0]
		peerID, err := getPeerID(hostname)
		if err == nil {
			err = client.PeerRemove(peerID)
		}
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"host":  hostname,
					"error": err.Error(),
				}).Error("peer remove failed")
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
		if verbose {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("peer status failed")
		}
		failure("Failed to get Peers list", err, 1)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Peer Addresses", "Online"})
	for _, peer := range peers {
		table.Append([]string{peer.ID.String(), peer.Name, strings.Join(peer.PeerAddresses, ","), formatBoolYesNo(peer.Online)})
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

// getPeerID return peerId of host
func getPeerID(host string) (string, error) {

	if uuid.Parse(host) != nil {
		return host, nil
	}
	// Get Peers list to find Peer ID
	peers, err := client.Peers()
	if err != nil {
		return "", err
	}

	peerID := ""

	hostinfo := strings.Split(host, ":")
	if len(hostinfo) == 1 {
		host = host + ":24008"
	}
	// Find Peer ID using available information
	for _, p := range peers {
		for _, h := range p.PeerAddresses {
			if h == host {
				peerID = p.ID.String()
				break
			}
		}
		// If already got Peer ID
		if peerID != "" {
			break
		}
	}

	if peerID == "" {
		return "", errors.New("Unable to find Peer ID")
	}

	return peerID, nil
}
