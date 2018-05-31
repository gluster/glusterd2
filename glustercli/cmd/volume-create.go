package cmd

import (
	"fmt"
	"strings"

	"github.com/gluster/glusterd2/pkg/api"
	smartvolapi "github.com/gluster/glusterd2/plugins/smartvol/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	volumeCreateHelpShort = "Create a Gluster volume"
	volumeCreateHelpLong  = "Create a Gluster volume of the requested type using the provided bricks. By default creates distribute volumes, unless specific volume type is requested by providing the relevant flags."
)

var (
	flagCreateStripeCount             int
	flagCreateReplicaCount            int
	flagCreateArbiterCount            int
	flagCreateDisperseCount           int
	flagCreateDisperseDataCount       int
	flagCreateDisperseRedundancyCount int
	flagCreateTransport               string
	flagCreateForce                   bool
	flagCreateAdvOpts                 bool
	flagCreateExpOpts                 bool
	flagCreateDepOpts                 bool
	flagCreateThinArbiter             string
	flagCreateVolumeOptions           []string

	flagCreateVolumeSize            string
	flagCreateDistributeCount       int
	flagCreateLimitPeers            []string
	flagCreateLimitZones            []string
	flagCreateExcludePeers          []string
	flagCreateExcludeZones          []string
	flagCreateSnapshotEnabled       bool
	flagCreateSnapshotReserveFactor float64 = 1
	flagCreateSubvolZoneOverlap     bool

	volumeCreateCmd = &cobra.Command{
		Use:   "create <volname> [<brick> [<brick>]...|--size <size>]",
		Short: volumeCreateHelpShort,
		Long:  volumeCreateHelpLong,
		Args:  cobra.MinimumNArgs(1),
		Run:   volumeCreateCmdRun,
	}
)

func init() {
	volumeCreateCmd.Flags().IntVar(&flagCreateStripeCount, "stripe", 0, "Stripe Count")
	volumeCreateCmd.Flags().IntVar(&flagCreateReplicaCount, "replica", 0, "Replica Count")
	volumeCreateCmd.Flags().IntVar(&flagCreateArbiterCount, "arbiter", 0, "Arbiter Count")
	volumeCreateCmd.Flags().StringVar(&flagCreateThinArbiter, "thin-arbiter", "",
		"Thin arbiter brick in the format <host>:<brick>[:<port>]. Port is optional and defaults to 24007")
	volumeCreateCmd.Flags().IntVar(&flagCreateDisperseCount, "disperse", 0, "Disperse Count")
	volumeCreateCmd.Flags().IntVar(&flagCreateDisperseDataCount, "disperse-data", 0, "Disperse Data Count")
	volumeCreateCmd.Flags().IntVar(&flagCreateDisperseRedundancyCount, "redundancy", 0, "Redundancy Count")
	volumeCreateCmd.Flags().StringVar(&flagCreateTransport, "transport", "tcp", "Transport")
	volumeCreateCmd.Flags().BoolVar(&flagCreateForce, "force", false, "Force")
	volumeCreateCmd.Flags().StringSliceVar(&flagCreateVolumeOptions, "options", nil,
		"Volume options in the format option:value,option:value")
	volumeCreateCmd.Flags().BoolVar(&flagCreateAdvOpts, "advanced", false, "Allow advanced options")
	volumeCreateCmd.Flags().BoolVar(&flagCreateExpOpts, "experimental", false, "Allow experimental options")
	volumeCreateCmd.Flags().BoolVar(&flagCreateDepOpts, "deprecated", false, "Allow deprecated options")
	volumeCreateCmd.Flags().BoolVar(&flagReuseBricks, "reuse-bricks", false, "Reuse bricks")
	volumeCreateCmd.Flags().BoolVar(&flagAllowRootDir, "allow-root-dir", false, "Allow root directory")
	volumeCreateCmd.Flags().BoolVar(&flagAllowMountAsBrick, "allow-mount-as-brick", false, "Allow mount as bricks")
	volumeCreateCmd.Flags().BoolVar(&flagCreateBrickDir, "create-brick-dir", false, "Create brick directory")

	// Smart Volume Flags
	volumeCreateCmd.Flags().StringVar(&flagCreateVolumeSize, "size", "", "Size of the Volume")
	volumeCreateCmd.Flags().IntVar(&flagCreateDistributeCount, "distribute", 1, "Distribute Count")
	volumeCreateCmd.Flags().StringSliceVar(&flagCreateLimitPeers, "limit-peers", nil, "Use bricks only from these Peers")
	volumeCreateCmd.Flags().StringSliceVar(&flagCreateLimitZones, "limit-zones", nil, "Use bricks only from these Zones")
	volumeCreateCmd.Flags().StringSliceVar(&flagCreateExcludePeers, "exclude-peers", nil, "Do not use bricks from these Peers")
	volumeCreateCmd.Flags().StringSliceVar(&flagCreateExcludeZones, "exclude-zones", nil, "Do not use bricks from these Zones")
	volumeCreateCmd.Flags().BoolVar(&flagCreateSnapshotEnabled, "enable-snapshot", false, "Enable Volume for Gluster Snapshot")
	volumeCreateCmd.Flags().Float64Var(&flagCreateSnapshotReserveFactor, "snapshot-reserve-factor", 1, "Snapshot Reserve Factor")
	volumeCreateCmd.Flags().BoolVar(&flagCreateSubvolZoneOverlap, "subvols-zones-overlap", false, "Brick belonging to other Sub volume can be created in the same zone")

	volumeCmd.AddCommand(volumeCreateCmd)
}

func smartVolumeCreate(cmd *cobra.Command, args []string) {
	size, err := sizeToMb(flagCreateVolumeSize)
	if err != nil {
		failure("Invalid Volume Size specified", nil, 1)
	}

	volname := args[0]

	req := smartvolapi.VolCreateReq{
		Name:                    volname,
		Transport:               flagCreateTransport,
		Size:                    size,
		ReplicaCount:            flagCreateReplicaCount,
		ArbiterCount:            flagCreateArbiterCount,
		DistributeCount:         flagCreateDistributeCount,
		DisperseCount:           flagCreateDisperseCount,
		DisperseDataCount:       flagCreateDisperseDataCount,
		DisperseRedundancyCount: flagCreateDisperseRedundancyCount,
		SnapshotEnabled:         flagCreateSnapshotEnabled,
		SnapshotReserveFactor:   flagCreateSnapshotReserveFactor,
		LimitPeers:              flagCreateLimitPeers,
		LimitZones:              flagCreateLimitZones,
		ExcludePeers:            flagCreateExcludePeers,
		ExcludeZones:            flagCreateExcludeZones,
		SubvolZonesOverlap:      flagCreateSubvolZoneOverlap,
		Force:                   flagCreateForce,
	}

	vol, err := client.SmartVolumeCreate(req)
	if err != nil {
		if verbose {
			log.WithFields(log.Fields{
				"volume": volname,
				"error":  err.Error(),
			}).Error("volume creation failed")
		}
		failure("Volume creation failed", err, 1)
	}
	fmt.Printf("%s Volume created successfully\n", vol.Name)
	fmt.Println("Volume ID: ", vol.ID)
}

func volumeCreateCmdRun(cmd *cobra.Command, args []string) {
	if flagCreateVolumeSize != "" {
		smartVolumeCreate(cmd, args)
		return
	}

	if len(args) < 2 {
		failure("Bricks not specified", nil, 1)
	}

	volname := args[0]
	bricks, err := bricksAsUUID(args[1:])
	if err != nil {
		if verbose {
			log.WithFields(log.Fields{
				"error":  err.Error(),
				"volume": volname,
			}).Error("error getting brick UUIDs")
		}
		failure("Error getting brick UUIDs", err, 1)
	}

	numBricks := len(bricks)
	subvols := []api.SubvolReq{}
	if flagCreateReplicaCount > 0 {
		// Replicate Volume Support

		if numBricks%flagCreateReplicaCount != 0 {
			failure("Invalid number of bricks specified", nil, 1)
		}

		numSubvols := numBricks / flagCreateReplicaCount

		for i := 0; i < numSubvols; i++ {
			idx := i * flagCreateReplicaCount

			// If Arbiter is set, set it as Brick Type for last brick
			if flagCreateArbiterCount > 0 {
				bricks[idx+flagCreateReplicaCount-1].Type = "arbiter"
			}

			subvols = append(subvols, api.SubvolReq{
				Type:         "replicate",
				Bricks:       bricks[idx : idx+flagCreateReplicaCount],
				ReplicaCount: flagCreateReplicaCount,
				ArbiterCount: flagCreateArbiterCount,
			})
		}
	} else if flagCreateDisperseCount > 0 || flagCreateDisperseDataCount > 0 || flagCreateDisperseRedundancyCount > 0 {
		subvolSize := 0
		if flagCreateDisperseCount > 0 {
			subvolSize = flagCreateDisperseCount
		} else if flagCreateDisperseDataCount > 0 && flagCreateDisperseRedundancyCount > 0 {
			subvolSize = flagCreateDisperseDataCount + flagCreateDisperseRedundancyCount
		}

		if subvolSize == 0 {
			failure("Invalid disperse-count/disperse-data/disperse-redundancy count", nil, 1)
		}

		if numBricks%subvolSize != 0 {
			failure("Invalid number of bricks specified", nil, 1)
		}

		numSubvols := numBricks / subvolSize

		for i := 0; i < numSubvols; i++ {
			idx := i * subvolSize

			subvols = append(subvols, api.SubvolReq{
				Type:               "disperse",
				Bricks:             bricks[idx : idx+subvolSize],
				DisperseCount:      flagCreateDisperseCount,
				DisperseData:       flagCreateDisperseDataCount,
				DisperseRedundancy: flagCreateDisperseRedundancyCount,
			})
		}
	} else {
		// Default Distribute Volume
		for _, b := range bricks {
			subvols = append(
				subvols,
				api.SubvolReq{
					Type:   "distribute",
					Bricks: []api.BrickReq{b},
				},
			)
		}
	}
	//set flags
	flags := make(map[string]bool)
	flags["reuse-bricks"] = flagReuseBricks
	flags["allow-root-dir"] = flagAllowRootDir
	flags["allow-mount-as-brick"] = flagAllowMountAsBrick
	flags["create-brick-dir"] = flagCreateBrickDir

	options := make(map[string]string)
	//set options
	if len(flagCreateVolumeOptions) != 0 {
		for _, opt := range flagCreateVolumeOptions {

			if len(strings.Split(opt, ":")) != 2 {
				fmt.Println("Error setting volume options")
				return
			}
			newopt := strings.Split(opt, ":")
			//check if empty option or value
			if len(strings.TrimSpace(newopt[0])) == 0 || len(strings.TrimSpace(newopt[1])) == 0 {
				fmt.Println("Error setting volume options ,contains empty option or value")
				return
			}
			options[newopt[0]] = newopt[1]
		}
	}

	req := api.VolCreateReq{
		Name:         volname,
		Subvols:      subvols,
		Force:        flagCreateForce,
		Options:      options,
		Advanced:     flagCreateAdvOpts,
		Experimental: flagCreateExpOpts,
		Deprecated:   flagCreateDepOpts,
		Flags:        flags,
	}

	// handle thin-arbiter
	if cmd.Flags().Changed("thin-arbiter") {
		if flagCreateReplicaCount != 2 {
			fmt.Println("Thin arbiter can only be enabled for replica count 2")
			return
		}
		if err := addThinArbiter(&req, cmd.Flag("thin-arbiter").Value.String()); err != nil {
			fmt.Println(err)
			return
		}
	}

	vol, err := client.VolumeCreate(req)
	if err != nil {
		if verbose {
			log.WithFields(log.Fields{
				"volume": volname,
				"error":  err.Error(),
			}).Error("volume creation failed")
		}
		failure("Volume creation failed", err, 1)
	}
	fmt.Printf("%s Volume created successfully\n", vol.Name)
	fmt.Println("Volume ID: ", vol.ID)
}

func addThinArbiter(req *api.VolCreateReq, thinArbiter string) error {

	s := strings.Split(thinArbiter, ":")
	if len(s) != 2 && len(s) != 3 {
		return fmt.Errorf("Thin arbiter brick must be of the form <host>:<brick> or <host>:<brick>:<port>")
	}

	// TODO: If required, handle this in a generic way, just like other
	// volume set options that we're going to allow to be set during
	// volume create.
	req.Options = map[string]string{
		"replicate.thin-arbiter": thinArbiter,
	}
	req.Advanced = true
	return nil
}
