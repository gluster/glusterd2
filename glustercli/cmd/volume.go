package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/olekukonko/tablewriter"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	helpVolumeExpandCmd = "Expand a Gluster Volume"
)

var (
	// Create Command Flags
	flagCreateCmdStripeCount       int
	flagCreateCmdReplicaCount      int
	flagCreateCmdArbiterCount      int
	flagCreateCmdDisperseCount     int
	flagCreateCmdDisperseDataCount int
	flagCreateCmdRedundancyCount   int
	flagCreateCmdTransport         string
	flagCreateCmdForce             bool

	// Start Command Flags
	flagStartCmdForce bool

	// Stop Command Flags
	flagStopCmdForce bool

	// Expand Command Flags
	flagExpandCmdForce bool
)

func init() {
	// Volume Create
	volumeCreateCmd.Flags().IntVarP(&flagCreateCmdStripeCount, "stripe", "", 0, "Stripe Count")
	volumeCreateCmd.Flags().IntVarP(&flagCreateCmdReplicaCount, "replica", "", 0, "Replica Count")
	volumeCreateCmd.Flags().IntVarP(&flagCreateCmdArbiterCount, "arbiter", "", 0, "Arbiter Count")
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

	// Volume Expand
	volumeExpandCmd.Flags().IntVarP(&flagCreateCmdReplicaCount, "replica", "", 0, "Replica Count")
	volumeExpandCmd.Flags().BoolVarP(&flagExpandCmdForce, "force", "f", false, "Force")
	volumeCmd.AddCommand(volumeExpandCmd)

	RootCmd.AddCommand(volumeCmd)
}

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: helpVolumeCmd,
}

func bricksAsUUID(bricks []string) ([]api.BrickReq, error) {

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
		var bs []api.BrickReq
		for _, b := range bricks {
			bData := strings.Split(b, ":")
			bs = append(bs, api.BrickReq{
				NodeID: bData[0],
				Path:   bData[1],
			})
		}
		return bs, nil
	}

	peers, err := client.Peers()
	if err != nil {
		return nil, err
	}

	var brickUUIDs []api.BrickReq

	for _, brick := range bricks {
		host := strings.Split(brick, ":")[0]
		path := strings.Split(brick, ":")[1]
		for _, peer := range peers {
			for _, addr := range peer.PeerAddresses {
				// TODO: Normalize presence/absence of port in peer address
				if strings.Split(addr, ":")[0] == strings.Split(host, ":")[0] {
					brickUUIDs = append(brickUUIDs, api.BrickReq{
						NodeID: peer.ID.String(),
						Path:   path,
					})
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
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		volname := cmd.Flags().Args()[0]
		bricks, err := bricksAsUUID(cmd.Flags().Args()[1:])
		if err != nil {
			log.WithFields(log.Fields{
				"error":  err.Error(),
				"volume": volname,
			}).Error("error getting brick UUIDs")
			failure("Error getting brick UUIDs", err, 1)
		}

		numBricks := len(bricks)
		subvols := []api.SubvolReq{}
		if flagCreateCmdReplicaCount > 0 {
			// Replicate Volume Support
			numSubvols := numBricks / flagCreateCmdReplicaCount

			for i := 0; i < numSubvols; i++ {
				idx := i * flagCreateCmdReplicaCount

				// If Arbiter is set, set it as Brick Type for last brick
				if flagCreateCmdArbiterCount > 0 {
					bricks[idx+flagCreateCmdReplicaCount-1].Type = "arbiter"
				}

				subvols = append(subvols, api.SubvolReq{
					Type:         "replicate",
					Bricks:       bricks[idx : idx+flagCreateCmdReplicaCount],
					ReplicaCount: flagCreateCmdReplicaCount,
					ArbiterCount: flagCreateCmdArbiterCount,
				})
			}
		} else {
			// Default Distribute Volume
			subvols = []api.SubvolReq{
				{
					Type:   "distribute",
					Bricks: bricks,
				},
			}
		}

		vol, err := client.VolumeCreate(api.VolCreateReq{
			Name:    volname,
			Subvols: subvols,
			Force:   flagCreateCmdForce,
		})
		if err != nil {
			log.WithFields(log.Fields{
				"volume": volname,
				"error":  err.Error(),
			}).Error("volume creation failed")
			failure("Volume creation failed", err, 1)
		}
		fmt.Printf("%s Volume created successfully\n", vol.Name)
		fmt.Println("Volume ID: ", vol.ID)
	},
}

var volumeStartCmd = &cobra.Command{
	Use:   "start [flags] <VOLNAME>",
	Short: helpVolumeStartCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		volname := cmd.Flags().Args()[0]
		err := client.VolumeStart(volname)
		if err != nil {
			log.WithFields(log.Fields{
				"volume": volname,
				"error":  err.Error(),
			}).Error("volume start failed")
			failure("volume start failed", err, 1)
		}
		fmt.Printf("Volume %s started successfully\n", volname)
	},
}

var volumeStopCmd = &cobra.Command{
	Use:   "stop [flags] <VOLNAME>",
	Short: helpVolumeStopCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		volname := cmd.Flags().Args()[0]
		err := client.VolumeStop(volname)
		if err != nil {
			log.WithFields(log.Fields{
				"volume": volname,
				"error":  err.Error(),
			}).Error("volume stop failed")
			failure("Volume stop failed", err, 1)
		}
		fmt.Printf("Volume %s stopped successfully\n", volname)
	},
}

var volumeDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: helpVolumeDeleteCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		volname := cmd.Flags().Args()[0]
		err := client.VolumeDelete(volname)
		if err != nil {
			log.WithFields(log.Fields{
				"volume": volname,
				"error":  err.Error(),
			}).Error("volume deletion failed")
			failure("Volume deletion failed", err, 1)
		}
		fmt.Printf("Volume %s deleted successfully\n", volname)
	},
}

var volumeGetCmd = &cobra.Command{
	Use:   "get",
	Short: helpVolumeGetCmd,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		volname := cmd.Flags().Args()[0]
		fmt.Println("GET:", volname)
	},
}

func volumeOptionJSONHandler(cmd *cobra.Command, volname string, options []string) error {
	vopt := make(map[string]string)
	for op, val := range options {
		if op%2 == 0 {
			vopt[val] = options[op+1]
		}
	}

	if volname == "all" {
		err := client.GlobalOptionSet(api.GlobalOptionReq{
			Options: vopt,
		})
		return err
	}

	err := client.VolumeSet(volname, api.VolOptionReq{
		Options: vopt,
	})
	return err
}

var volumeSetCmd = &cobra.Command{
	Use:   "set",
	Short: helpVolumeSetCmd,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		if (len(cmd.Flags().Args())-1)%2 == 0 {
			volname := cmd.Flags().Args()[0]
			options := cmd.Flags().Args()[1:]
			err := volumeOptionJSONHandler(cmd, volname, options)
			if err != nil {
				log.WithFields(log.Fields{
					"volume": volname,
					"error":  err.Error(),
				}).Error("volume option set failed")
				failure("Volume option set failed", err, 1)
			} else {
				fmt.Printf("Options set successfully for %s volume\n", volname)
			}
		} else {
			fmt.Println("Incorrect volume options")
		}
	},
}

var volumeResetCmd = &cobra.Command{
	Use:   "reset",
	Short: helpVolumeResetCmd,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		volname := cmd.Flags().Args()[0]
		fmt.Println("RESET:", volname)
	},
}

func volumeInfoDisplayNumbricks(vol api.VolumeGetResp) {

	var DistCount = len(vol.Subvols)
	// TODO: Assumption as all subvol types are same
	var RepCount = vol.Subvols[0].ReplicaCount
	var ArbCount = vol.Subvols[0].ArbiterCount
	numBricks := 0
	for _, subvol := range vol.Subvols {
		numBricks += len(subvol.Bricks)
	}

	if DistCount > 1 && vol.Subvols[0].Type == api.SubvolReplicate {
		if ArbCount == 1 {
			fmt.Printf("Number of Bricks: %d x (%d + %d) = %d\n", DistCount, RepCount-1, ArbCount, numBricks)
		} else {
			fmt.Printf("Number of Bricks: %d x %d = %d\n", DistCount, RepCount, numBricks)
		}
	} else {
		fmt.Println("Number of Bricks:", numBricks)
	}
}

func volumeInfoDisplay(vol api.VolumeGetResp) {

	fmt.Println()
	fmt.Println("Volume Name:", vol.Name)
	fmt.Println("Type:", vol.Type)
	fmt.Println("Volume ID:", vol.ID)
	fmt.Println("State:", vol.State)
	fmt.Println("Transport-type:", vol.Transport)
	volumeInfoDisplayNumbricks(vol)
	for sIdx, subvol := range vol.Subvols {
		for bIdx, brick := range subvol.Bricks {
			if brick.Type == api.Arbiter {
				fmt.Printf("Brick%d: %s:%s (arbiter)\n", sIdx+bIdx+1, brick.NodeID, brick.Path)
			} else {
				fmt.Printf("Brick%d: %s:%s\n", sIdx+bIdx+1, brick.NodeID, brick.Path)
			}
		}
	}
	return
}
func volumeInfoHandler2(cmd *cobra.Command, isInfo bool) error {
	var vols api.VolumeListResp
	var err error
	volname := ""
	if len(cmd.Flags().Args()) > 0 {
		volname = cmd.Flags().Args()[0]
	}
	if volname == "" {
		vols, err = client.Volumes("")
	} else {
		vols, err = client.Volumes(volname)
	}

	if err != nil {
		return err
	}

	if isInfo {
		for _, vol := range vols {
			volumeInfoDisplay(vol)
		}
	} else {
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
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		err := volumeInfoHandler2(cmd, true)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error getting volumes list")
			failure("Error getting Volumes list", err, 1)
		}
	},
}

var volumeListCmd = &cobra.Command{
	Use:   "list",
	Short: helpVolumeListCmd,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		err := volumeInfoHandler2(cmd, false)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error getting volumes list")
			failure("Error getting Volumes list", err, 1)
		}
	},
}

func volumeStatusDisplay(vol api.BricksStatusResp) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Brick ID", "Host", "Path", "Online", "Port", "Pid"})
	for _, b := range vol {
		table.Append([]string{b.Info.ID.String(), b.Info.Hostname, b.Info.Path,
			strconv.FormatBool(b.Online), strconv.Itoa(b.Port), strconv.Itoa(b.Pid)})
	}
	table.Render()
}

func volumeStatusHandler(cmd *cobra.Command) error {
	var vol api.BricksStatusResp
	var err error
	volname := ""
	if len(cmd.Flags().Args()) > 0 {
		volname = cmd.Flags().Args()[0]
	}
	if volname == "" {
		var volList api.VolumeListResp
		volList, err = client.Volumes("")
		for _, volume := range volList {
			vol, err = client.BricksStatus(volume.Name)
			fmt.Println("Volume :", volume.Name)
			if err == nil {
				volumeStatusDisplay(vol)
			} else {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error getting volume status")
				failure("Error getting Volume status", err, 1)
			}
		}
	} else {
		vol, err = client.BricksStatus(volname)
		fmt.Println("Volume :", volname)
		if err == nil {
			volumeStatusDisplay(vol)
		}
	}
	return err
}

var volumeStatusCmd = &cobra.Command{
	Use:   "status",
	Short: helpVolumeStatusCmd,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		err := volumeStatusHandler(cmd)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error getting volume status")
			failure("Error getting Volume status", err, 1)
		}
	},
}

var volumeExpandCmd = &cobra.Command{
	Use:   "add-brick",
	Short: helpVolumeExpandCmd,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		volname := cmd.Flags().Args()[0]
		bricks, err := bricksAsUUID(cmd.Flags().Args()[1:])
		if err != nil {
			log.WithFields(log.Fields{
				"volume": volname,
				"error":  err.Error(),
			}).Error("error getting brick UUIDs")
			failure("Error getting brick UUIDs", err, 1)
		}
		vol, err := client.VolumeExpand(volname, api.VolExpandReq{
			ReplicaCount: flagCreateCmdReplicaCount,
			Bricks:       bricks, // string of format <UUID>:<path>
			Force:        flagExpandCmdForce,
		})
		if err != nil {
			log.WithFields(log.Fields{
				"volume": volname,
				"error":  err.Error(),
			}).Error("volume expansion failed")
			failure("Addition of brick failed", err, 1)
		}
		fmt.Printf("%s Volume expanded successfully\n", vol.Name)
	},
}
