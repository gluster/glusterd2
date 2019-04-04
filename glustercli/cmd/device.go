package cmd

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/plugins/device/api"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpDeviceCmd     = "Gluster Devices Management"
	helpDeviceAddCmd  = "Add device"
	helpDeviceInfoCmd = "Get device info"
)

var flagDeviceAddProvisioner string

func init() {
	deviceAddCmd.Flags().StringVar(&flagDeviceAddProvisioner, "provisioner", "lvm", "Provisioner Type(lvm, loop)")
	deviceCmd.AddCommand(deviceAddCmd)
	deviceCmd.AddCommand(deviceInfoCmd)
}

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: helpDeviceCmd,
}

var deviceInfoCmd = &cobra.Command{
	Use:   "info <PeerID>",
	Short: helpDeviceInfoCmd,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var peerID string
		if len(args) == 1 {
			peerID = args[0]
		}

		if peerID == "" {
			peers, err := client.Peers()
			if err != nil {
				if GlobalFlag.Verbose {
					log.WithError(err).Error("peer list failed")
				}
				failure("Failed to get peer list", err, 1)
			}
			for _, peer := range peers {
				peerID := peer.ID.String()
				deviceList, err := client.DeviceList(peerID, "")
				if err != nil {
					if GlobalFlag.Verbose {
						log.WithError(err).Error("device list failed")
					}
					failure("Failed to get device list", err, 1)
				}
				if len(deviceList) == 0 {
					continue
				}
				deviceListDisplay(peerID, peer.Name, deviceList)
			}
			return
		}

		peerInfo, err := client.GetPeer(peerID)
		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).Error("peer list failed")
			}
			failure("Failed to get peer list", err, 1)
		}

		deviceList, err := client.DeviceList(peerID, "")
		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).WithField("peerid", peerID).Error("device list failed")
			}
			failure("Device list failed", err, 1)
		}
		if len(deviceList) == 0 {
			fmt.Println("No devices are associated with given peer")
			return
		}
		deviceListDisplay(peerID, peerInfo.Name, deviceList)
	},
}

func deviceListDisplay(peerID, peerName string, deviceList []api.Info) {
	fmt.Println()
	fmt.Println("Peer Name:", peerName)
	fmt.Println("Peer ID:", peerID)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Device", "State", "Total Size", "Free Size", "Used Size", "Used %"})
	for _, d := range deviceList {
		var usedPer float64
		if d.UsedSize > 0 {
			usedPer = float64(d.UsedSize) / float64(d.TotalSize) * 100
		}

		table.Append([]string{d.Device, d.State, humanReadable(d.TotalSize),
			humanReadable(d.AvailableSize), humanReadable(d.UsedSize), fmt.Sprintf("%.2f", usedPer)})
	}
	table.Render()
}

var deviceAddCmd = &cobra.Command{
	Use:   "add <PeerID> <DEVICE>",
	Short: helpDeviceAddCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		peerid := args[0]
		devname := args[1]

		_, err := client.DeviceAdd(peerid, devname, flagDeviceAddProvisioner)

		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).WithFields(log.Fields{
					"device": devname,
					"peerid": peerid,
				}).Error("device add failed")
			}
			failure("Device add failed", err, 1)
		}
		fmt.Println("Device add successful")
	},
}
