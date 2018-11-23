package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpDeviceCmd    = "Gluster Devices Management"
	helpDeviceAddCmd = "add device"
)

func init() {
	deviceCmd.AddCommand(deviceAddCmd)
}

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: helpDeviceCmd,
}

var deviceAddCmd = &cobra.Command{
	Use:   "add <PeerID> <DEVICE>",
	Short: helpDeviceAddCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		peerid := args[0]
		devname := args[1]

		_, err := client.DeviceAdd(peerid, devname)

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
