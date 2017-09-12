package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/olekukonko/tablewriter"
)

const (
	helpVolumeCmd       = "Gluster Volume Management"
	helpVolumeCreateCmd = "Create a Gluster Volume"
	helpVolumeStartCmd  = "Start a Gluster Volume"
	helpVolumeStopCmd   = "Stop a Gluster Volume"
	helpVolumeDeleteCmd = "Delete a Gluster Volume"
	helpVolumeGetCmd    = "Get Gluster Volume Options"
	helpVolumeSetCmd    = "Set a Gluster Volume Option"
	helpVolumeResetCmd  = "Reset a Gluster Volume Option"
	helpVolumeInfoCmd   = "Get Gluster Volume Info"
	helpVolumeListCmd   = "List all Gluster Volumes"
	helpVolumeStatusCmd = "Get Gluster Volume Status"
)

var (
	// Create Command Flags
	flagCreateCmdStripeCount       int
	flagCreateCmdReplicaCount      int
	flagCreateCmdDisperseCount     int
	flagCreateCmdDisperseDataCount int
	flagCreateCmdRedundancyCount   int
	flagCreateCmdTransport         string
	flagCreateCmdForce             bool

	// Start Command Flags
	flagStartCmdForce bool

	// Stop Command Flags
	flagStopCmdForce bool
)

func init() {
	// Volume Create
	volumeCreateCmd.Flags().IntVarP(&flagCreateCmdStripeCount, "stripe", "", 0, "Stripe Count")
	volumeCreateCmd.Flags().IntVarP(&flagCreateCmdReplicaCount, "replica", "", 0, "Replica Count")
	volumeCreateCmd.Flags().IntVarP(&flagCreateCmdDisperseCount, "disperse", "", 0, "Disperse Count")
	volumeCreateCmd.Flags().IntVarP(&flagCreateCmdDisperseDataCount, "disperse-data", "", 0, "Disperse Data Count")
	volumeCreateCmd.Flags().IntVarP(&flagCreateCmdRedundancyCount, "redundancy", "", 0, "Redundancy Count")
	volumeCreateCmd.Flags().StringVarP(&flagCreateCmdTransport, "transport", "", "tcp", "Transport")
	volumeCreateCmd.Flags().BoolVarP(&flagCreateCmdForce, "force", "f", false, "Force")
	volumeCmd.AddCommand(volumeCreateCmd)

	// Volume Start
	volumeStartCmd.Flags().BoolVarP(&flagStartCmdForce, "force", "f", false, "Force")
	volumeCmd.AddCommand(volumeStartCmd)

	// Volume Stop
	volumeStopCmd.Flags().BoolVarP(&flagStopCmdForce, "force", "f", false, "Force")
	volumeCmd.AddCommand(volumeStopCmd)

	// Volume Delete
	volumeCmd.AddCommand(volumeDeleteCmd)

	volumeCmd.AddCommand(volumeGetCmd)
	volumeCmd.AddCommand(volumeSetCmd)
	volumeCmd.AddCommand(volumeResetCmd)
	volumeCmd.AddCommand(volumeInfoCmd)
	volumeCmd.AddCommand(volumeStatusCmd)
	volumeCmd.AddCommand(volumeListCmd)
	RootCmd.AddCommand(volumeCmd)
}

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: helpVolumeCmd,
}

func bricksAsUUID(bricks []string) ([]string, error) {

	// validate if <host> in <host>:<path> is already UUID
	validUUIDs := 0
	for _, brick := range bricks {
		host := strings.Split(brick, ":")[0]
		if uuid.Parse(host) == nil {
			break
		}
		validUUIDs++
	}

	if validUUIDs == len(bricks) {
		// bricks are already of the format <uuid>:<path>
		return bricks, nil
	}

	peers, err := client.Peers()
	if err != nil {
		return nil, err
	}

	var brickUUIDs []string

	for _, brick := range bricks {
		host := strings.Split(brick, ":")[0]
		path := strings.Split(brick, ":")[1]
		for _, peer := range peers {
			for _, addr := range peer.Addresses {
				// TODO: Normalize presence/absence of port in peer address
				if strings.Split(addr, ":")[0] == strings.Split(host, ":")[0] {
					brickUUIDs = append(brickUUIDs, peer.ID.String()+":"+path)
				}
			}
		}
	}

	if len(brickUUIDs) != len(bricks) {
		return brickUUIDs, errors.New("could not find UUIDs of bricks specified")
	}

	return brickUUIDs, nil
}

var volumeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: helpVolumeCreateCmd,
	Run: func(cmd *cobra.Command, args []string) {
		validateNArgs(cmd, 2, 0)
		volname := cmd.Flags().Args()[0]
		bricks, err := bricksAsUUID(cmd.Flags().Args()[1:])
		if err != nil {
			log.WithField("volume", volname).Println("volume creation failed")
			failure(fmt.Sprintf("Error getting brick UUIDs: %s", err.Error()), 1)
		}
		vol, err := client.VolumeCreate(api.VolCreateReq{
			Name:    volname,
			Bricks:  bricks, // string of format <UUID>:<path>
			Replica: flagCreateCmdReplicaCount,
			Force:   flagCreateCmdForce,
		})
		if err != nil {
			log.WithField("volume", volname).Println("volume creation failed")
			failure(fmt.Sprintf("Volume creation failed with %s", err.Error()), 1)
		}
		fmt.Printf("%s Volume created successfully\n", vol.Name)
		fmt.Println("Volume ID: ", vol.ID)
	},
}

var volumeStartCmd = &cobra.Command{
	Use:   "start [flags] <VOLNAME>",
	Short: helpVolumeStartCmd,
	Run: func(cmd *cobra.Command, args []string) {
		validateNArgs(cmd, 1, 1)
		volname := cmd.Flags().Args()[0]
		err := client.VolumeStart(volname)
		if err != nil {
			log.WithField("volume", volname).Println("volume start failed")
			failure(fmt.Sprintf("volume start failed with: %s", err.Error()), 1)
		}
		fmt.Printf("Volume %s started successfully\n", volname)
	},
}

var volumeStopCmd = &cobra.Command{
	Use:   "stop [flags] <VOLNAME>",
	Short: helpVolumeStopCmd,
	Run: func(cmd *cobra.Command, args []string) {
		validateNArgs(cmd, 1, 1)
		volname := cmd.Flags().Args()[0]
		err := client.VolumeStop(volname)
		if err != nil {
			log.WithField("volume", volname).Println("volume stop failed")
			failure(fmt.Sprintf("volume stop failed with: %s", err.Error()), 1)
		}
		fmt.Printf("Volume %s stopped successfully\n", volname)
	},
}

var volumeDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: helpVolumeDeleteCmd,
	Run: func(cmd *cobra.Command, args []string) {
		validateNArgs(cmd, 1, 1)
		volname := cmd.Flags().Args()[0]
		err := client.VolumeDelete(volname)
		if err != nil {
			log.WithField("volume", volname).Println("volume deletion failed")
			failure(fmt.Sprintf("volume deletion failed with: %s", err.Error()), 1)
		}
		fmt.Printf("Volume %s deleted successfully\n", volname)
	},
}

var volumeGetCmd = &cobra.Command{
	Use:   "get",
	Short: helpVolumeGetCmd,
	Run: func(cmd *cobra.Command, args []string) {
		validateNArgs(cmd, 1, 2)
		volname := cmd.Flags().Args()[0]
		fmt.Println("GET:", volname)
	},
}

var volumeSetCmd = &cobra.Command{
	Use:   "set",
	Short: helpVolumeSetCmd,
	Run: func(cmd *cobra.Command, args []string) {
		validateNArgs(cmd, 3, 3)
		volname := cmd.Flags().Args()[0]
		fmt.Println("SET:", volname)
	},
}

var volumeResetCmd = &cobra.Command{
	Use:   "reset",
	Short: helpVolumeResetCmd,
	Run: func(cmd *cobra.Command, args []string) {
		validateNArgs(cmd, 1, 2)
		volname := cmd.Flags().Args()[0]
		fmt.Println("RESET:", volname)
	},
}

func volumeInfoHandler2(cmd *cobra.Command, isInfo bool) error {
	var vols []api.Volinfo
	var err error
	validateNArgs(cmd, 0, 1)
	volname := ""
	if len(cmd.Flags().Args()) > 0 {
		volname = cmd.Flags().Args()[0]
	}
	if volname == ""{
		vols, err = client.Volumes("")
	}else{
		vols, err = client.Volumes(volname)
	}
	if isInfo{
		for _, vol := range vols {
			fmt.Println("Volume Name: ", vol.Name)
			fmt.Println("Type: ", vol.Type)
			fmt.Println("Volume ID: ", vol.ID)
			fmt.Println("Status: ", vol.Status)
			fmt.Println("Transport-type: ", vol.Transport)
			fmt.Println("Number of Bricks: ", len(vol.Bricks))
			fmt.Println("Bricks:")
			for i,brick := range vol.Bricks{
				fmt.Printf("Brick%d: %s:%s\n", i+1, brick.Hostname, brick.Path)
			}
		}
	}else{
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name"})
		for _, vol := range vols {
			table.Append([]string{vol.ID.String(), vol.Name})
		}
		table.Render()
	}
	return err
}

var volumeInfoCmd = &cobra.Command{
	Use:   "info",
	Short: helpVolumeInfoCmd,
	Run: func(cmd *cobra.Command, args []string) {
		err := volumeInfoHandler2(cmd, true)
		if err != nil {
			log.Println("No such volume present")
			failure(fmt.Sprintf("Error getting Volumes list %s", err.Error()), 1)
		}
	},
}

var volumeListCmd = &cobra.Command{
	Use:   "list",
	Short: helpVolumeListCmd,
	Run: func(cmd *cobra.Command, args []string) {
		err := volumeInfoHandler2(cmd, false)
		if err != nil {
			log.Println("No volumes present")
			failure(fmt.Sprintf("Error getting Volumes list %s", err.Error()), 1)
		}
	},
}

var volumeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: helpVolumeStatusCmd,
	Run: func(cmd *cobra.Command, args []string) {
		validateNArgs(cmd, 0, 1)
		volname := "all"
		if len(cmd.Flags().Args()) > 0 {
			volname = cmd.Flags().Args()[0]
		}
		fmt.Println("STATUS:", volname)
	},
}
