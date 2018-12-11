package cmd

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Profile Info Flags
	flagProfileInfoPeek            bool
	flagProfileInfoIncremental     bool
	flagProfileInfoIncrementalPeek bool
	flagProfileInfoCumulative      bool
	flagProfileInfoClear           bool
)

var volumeProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Gluster volume profile",
	Long:  "Gluster Volume Profile retrieves info on stats like latency, no of fops performed on the volume etc",
	Args:  cobra.ExactArgs(2),
}

func init() {
	// Volume Profile Info
	volumeProfileInfoCmd.Flags().BoolVar(&flagProfileInfoPeek, "peek", false, "Volume Profile Info Peek")
	volumeProfileInfoCmd.Flags().BoolVar(&flagProfileInfoIncremental, "incremental", false, "Volume Profile Info Incremental")
	volumeProfileInfoCmd.Flags().BoolVar(&flagProfileInfoIncrementalPeek, "incremental-peek", false, "Volume Profile Info Incremental Peek")
	volumeProfileInfoCmd.Flags().BoolVar(&flagProfileInfoCumulative, "cumulative", false, "Volume Profile Info Cumulative")
	volumeProfileInfoCmd.Flags().BoolVar(&flagProfileInfoClear, "clear", false, "Volume Profile Info Clear")

	volumeProfileCmd.AddCommand(volumeProfileInfoCmd)

	volumeCmd.AddCommand(volumeProfileCmd)
}

var volumeProfileInfoCmd = &cobra.Command{
	Use:   "info <volname> [--peek|--incremental|--incremental-peek|--cumulative|--clear]",
	Short: "Volume Profile Info",
	Long:  "Volume Profile Info retrieves stats like latency, read/write bytes, no of fops performed on the volume etc",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var volumeProfileInfo []api.BrickProfileInfo
		volname := args[0]
		option := "info"
		if flagProfileInfoPeek {
			option = "info-peek"
		} else if flagProfileInfoIncremental {
			option = "info-incremental"
		} else if flagProfileInfoIncrementalPeek {
			option = "info-incremental-peek"
		} else if flagProfileInfoCumulative {
			option = "info-cumulative"
		} else if flagProfileInfoClear {
			option = "info-clear"
		}

		volumeProfileInfo, err = client.VolumeProfileInfo(volname, option)
		if err != nil {
			log.WithError(err).WithField("volname", volname).Error("failed to get volume profile info")
			failure(fmt.Sprintf("Failed to get volume profile info for volume %s\n", volname), err, 1)
		}
		// Iterate over all bricks
		for index := range volumeProfileInfo {

			// Display Cumulative Stats
			table := tablewriter.NewWriter(os.Stdout)
			fmt.Printf("Brick: %s\n\n", volumeProfileInfo[index].BrickName)
			if volumeProfileInfo[index].CumulativeStats.Interval != "" {
				fmt.Printf("Cumulative Stats: \n")
			}
			table.SetHeader([]string{"%-Latency", "AvgLatency", "MinLatency", "MaxLatency", "No. Of Calls", "FOP"})
			if len(volumeProfileInfo[index].CumulativeStats.StatsInfo) != 0 {
				// Iterate over stats of fop in Cumulative stats, key being the FOP name
				for key := range volumeProfileInfo[index].CumulativeStats.StatsInfo {
					table.Append([]string{volumeProfileInfo[index].CumulativeStats.StatsInfo[key]["%-latency"],
						volumeProfileInfo[index].CumulativeStats.StatsInfo[key]["avglatency"],
						volumeProfileInfo[index].CumulativeStats.StatsInfo[key]["minlatency"],
						volumeProfileInfo[index].CumulativeStats.StatsInfo[key]["maxlatency"],
						volumeProfileInfo[index].CumulativeStats.StatsInfo[key]["hits"],
						key})
				}
				table.Render()
			}
			if volumeProfileInfo[index].CumulativeStats.Duration != "" {
				fmt.Printf("Duration: %s seconds\n", volumeProfileInfo[index].CumulativeStats.Duration)
				fmt.Printf("Data Read: %s bytes\n", volumeProfileInfo[index].CumulativeStats.DataRead)
				fmt.Printf("Data Write: %s bytes\n\n\n", volumeProfileInfo[index].CumulativeStats.DataWrite)
			}
			fmt.Printf("\n\n")

			// Display Interval Stats
			table = tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"%-Latency", "AvgLatency", "MinLatency", "MaxLatency", "No. Of Calls", "FOP"})
			if volumeProfileInfo[index].IntervalStats.Interval != "" {
				fmt.Printf("Interval %s Stats: \n\n", volumeProfileInfo[index].IntervalStats.Interval)
			}
			if len(volumeProfileInfo[index].IntervalStats.StatsInfo) != 0 {
				// Iterate over stats of fop in Cumulative stats, key being the FOP name
				for key := range volumeProfileInfo[index].IntervalStats.StatsInfo {
					table.Append([]string{volumeProfileInfo[index].IntervalStats.StatsInfo[key]["%-latency"],
						volumeProfileInfo[index].IntervalStats.StatsInfo[key]["avglatency"],
						volumeProfileInfo[index].IntervalStats.StatsInfo[key]["minlatency"],
						volumeProfileInfo[index].IntervalStats.StatsInfo[key]["maxlatency"],
						volumeProfileInfo[index].IntervalStats.StatsInfo[key]["hits"],
						key})
				}
				table.Render()
			}
			if volumeProfileInfo[index].IntervalStats.Duration != "" {
				fmt.Printf("Duration: %s seconds\n", volumeProfileInfo[index].IntervalStats.Duration)
				fmt.Printf("Data Read: %s bytes\n", volumeProfileInfo[index].IntervalStats.DataRead)
				fmt.Printf("Data Write: %s bytes\n", volumeProfileInfo[index].IntervalStats.DataWrite)
			}
			fmt.Printf("\n\n")
		}
	},
}
