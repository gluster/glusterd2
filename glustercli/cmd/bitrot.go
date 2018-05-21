package cmd

import (
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpBitrotCmd               = "Gluster Bitrot"
	helpBitrotEnableCmd         = "Enable Bitrot"
	helpBitrotDisableCmd        = "Disable Bitrot"
	helpBitrotScrubThrottleCmd  = "Configure Scrub Throttle"
	helpBitrotScrubFrequencyCmd = "Configure Scrub Frequency"
	helpBitrotScrubCmd          = "Bitrot Scrub Command"
)

const (
	scrubPause    = "pause"
	scrubResume   = "resume"
	scrubOndemand = "ondemand"
	scrubStatus   = "status"
)

func init() {
	// Bitrot Enable
	bitrotCmd.AddCommand(bitrotEnableCmd)

	// Bitrot Disable
	bitrotCmd.AddCommand(bitrotDisableCmd)

	// Configure scrub throttle
	bitrotCmd.AddCommand(bitrotScrubThrottleCmd)

	// Configure scrub frequency
	bitrotCmd.AddCommand(bitrotScrubFrequencyCmd)

	// Bitrot scrub command
	bitrotCmd.AddCommand(bitrotScrubCmd)

	RootCmd.AddCommand(bitrotCmd)

}

var bitrotCmd = &cobra.Command{
	Use:   "bitrot",
	Short: helpBitrotCmd,
}

var bitrotEnableCmd = &cobra.Command{
	Use:   "enable <volname>",
	Short: helpBitrotEnableCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		volname := args[0]
		err := client.BitrotEnable(volname)
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"volume": volname,
					"error":  err.Error(),
				}).Error("failed to enable bitrot")
			}
			failure(fmt.Sprintf("Failed to enable bitrot for volume %s\n", volname), err, 1)
		}
		fmt.Printf("Bitrot enabled successfully for volume %s\n", volname)
	},
}

var bitrotDisableCmd = &cobra.Command{
	Use:   "disable <volname>",
	Short: helpBitrotDisableCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		volname := args[0]

		err := client.BitrotDisable(volname)
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"volume": volname,
					"error":  err.Error(),
				}).Error("failed to disable bitrot")
			}
			failure(fmt.Sprintf("Failed to disable bitrot for volume %s\n", volname), err, 1)
		}
		fmt.Printf("Bitrot disabled successfully for volume '%s'\n", volname)
	},
}

var bitrotScrubThrottleCmd = &cobra.Command{
	Use:   "scrub-throttle <volname> {lazy|normal|aggressive}",
	Short: helpBitrotScrubThrottleCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		volname := args[0]
		var option []string
		option = append(option, "bit-rot.scrub-throttle")
		option = append(option, args[1])

		// Set Option set flag to advanced
		flagSetAdv = true
		err := volumeOptionJSONHandler(cmd, volname, option)
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"volume": volname,
					"value":  args[1],
					"error":  err.Error(),
				}).Error("failed to set scrub-throttle")
			}
			failure(fmt.Sprintf("Failed to set bitrot scrub throttle to %s for volume %s", args[1], volname), err, 1)
		}
		fmt.Printf("Bitrot scrub throttle set successfully to %s for volume %s\n", args[1], volname)
	},
}

var bitrotScrubFrequencyCmd = &cobra.Command{
	Use:   "scrub-freq <volname> {hourly|daily|weekly|biweekly|monthly}",
	Short: helpBitrotScrubFrequencyCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		volname := args[0]
		var option []string
		option = append(option, "bit-rot.scrub-freq")
		option = append(option, args[1])

		// Set Option set flag to advanced
		flagSetAdv = true
		err := volumeOptionJSONHandler(cmd, volname, option)
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"volume": volname,
					"value":  args[1],
					"error":  err.Error(),
				}).Error("failed to set scrub-frequency")
			}
			failure(fmt.Sprintf("Failed to set bitrot scrub frequency to %s for volume %s", args[1], volname), err, 1)
		}
		fmt.Printf("Bitrot scrub frequency is set successfully to %s for volume %s\n", args[1], volname)
	},
}

var bitrotScrubCmd = &cobra.Command{
	Use:   "scrub <volname> {pause|resume|status|ondemand}",
	Short: helpBitrotScrubCmd,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		volname := args[0]
		var option []string

		switch scrubCmd := args[1]; scrubCmd {
		case scrubPause, scrubResume:
			option = append(option, "bit-rot.scrub-state")
			option = append(option, args[1])

			// Set Option set flag to advanced
			flagSetAdv = true
			err := volumeOptionJSONHandler(cmd, volname, option)
			if err != nil {
				if verbose {
					log.WithFields(log.Fields{
						"volume": volname,
						"value":  args[1],
						"error":  err.Error(),
					}).Error("Bitrot scrub", scrubCmd, "command failed")
				}
				failure(fmt.Sprintf("Failed to %s bitrot scrub for volume %s", args[1], volname), err, 1)
			}
			fmt.Printf("Bitrot scrub %s is successful for volume %s\n", args[1], volname)

		case scrubStatus:
			scrubStatus, err := client.BitrotScrubStatus(volname)
			if err != nil {
				if verbose {
					log.WithFields(log.Fields{
						"volume": volname,
						"error":  err.Error(),
					}).Error("failed to get bitrot scrub status")
				}
				failure(fmt.Sprintf("Failed to get bitrot scrub status for volume %s\n", volname), err, 1)
			}
			fmt.Println()
			fmt.Printf("Volume: %s\n", scrubStatus.Volume)
			fmt.Printf("Scrub state: %s\n", scrubStatus.State)
			fmt.Printf("Scrub impact: %s\n", scrubStatus.Throttle)
			fmt.Printf("Scrub frequency: %s\n", scrubStatus.Frequency)
			fmt.Printf("Bitd log file: %s\n", scrubStatus.BitdLogFile)
			fmt.Printf("Scrubber log file: %s\n\n", scrubStatus.ScrubLogFile)

			for _, nodeInfo := range scrubStatus.Nodes {
				// TODO: Convert node id into hostname
				fmt.Printf("Node: %s\n", nodeInfo.Node)
				fmt.Printf("==========================================\n")
				fmt.Printf("Number of scrubbed files: %s\n", nodeInfo.NumScrubbedFiles)
				fmt.Printf("Number of skipped files: %s\n", nodeInfo.NumSkippedFiles)
				fmt.Printf("Last completed scrub time: %s\n", nodeInfo.LastScrubCompletedTime)

				/* Printing last scrub duration time in human readable form*/
				scrubTime, err := strconv.Atoi(nodeInfo.LastScrubDuration)
				if err != nil {
					failure(fmt.Sprintf("Failed to parse bitrot scrub status for volume %s\n", volname), err, 1)
				}
				seconds := scrubTime % 60
				minutes := (scrubTime / 60) % 60
				hours := (scrubTime / 3600) % 24
				days := scrubTime / 86400

				fmt.Printf("Duration of last scrub (days:hrs:mins:secs): %d:%d:%d:%d\n", days, hours, minutes, seconds)
				fmt.Println()
				fmt.Printf("Number of corrupted objects: %s\n", nodeInfo.ErrorCount)
				fmt.Println("Corrupted object's GFID:")
				for _, corruptedObject := range nodeInfo.CorruptedObjects {
					fmt.Println(corruptedObject)
				}
				fmt.Println()
			}

		case scrubOndemand:
			err := client.BitrotScrubOndemand(volname)
			if err != nil {
				if verbose {
					log.WithFields(log.Fields{
						"volume": volname,
						"error":  err.Error(),
					}).Error("failed to start bitrot scrub on demand")
				}
				failure(fmt.Sprintf("Failed to start bitrot scrub on demand for volume %s\n", volname), err, 1)
			}
			fmt.Printf("Bitrot scrub on demand started successfully for volume %s\n", volname)
		default:
			failure(fmt.Sprintf(
				"Invalid scrub value: %s\nUsage: glustercli bitrot scrub <volname> {pause|resume|status|ondemand}",
				scrubCmd), nil, 1)
		}

	},
}
