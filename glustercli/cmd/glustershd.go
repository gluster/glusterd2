package cmd

import (
	"errors"
	"fmt"
	"strings"

	glustershdapi "github.com/gluster/glusterd2/plugins/glustershd/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Heal Info Flags
	flagSummaryInfo    bool
	flagSplitBrainInfo bool

	// Split Brain Flags
	flagSplitBrainBiggerFile  bool
	flagSplitBrainLatestMtime bool
	flagSplitBrainSourceBrick string
	flagFileName              string
)

var selfHealCmd = &cobra.Command{
	Use:   "heal",
	Short: "Gluster Self Heal",
	Args:  cobra.MinimumNArgs(2),
}

func init() {
	// Self Heal Info
	selfHealInfoCmd.Flags().BoolVar(&flagSummaryInfo, "info-summary", false, "Heal Info Summary")
	selfHealInfoCmd.Flags().BoolVar(&flagSplitBrainInfo, "split-brain-info", false, "Heal Split Brain Info")
	selfHealCmd.AddCommand(selfHealInfoCmd)
	selfHealCmd.AddCommand(selfHealIndexCmd)
	selfHealCmd.AddCommand(selfHealFullCmd)

	selfHealSplitBrainCmd.Flags().BoolVar(&flagSplitBrainBiggerFile, "bigger-file", false, "Use bigger-file to resolve split-brain")
	selfHealSplitBrainCmd.Flags().BoolVar(&flagSplitBrainLatestMtime, "latest-mtime", false, "Use latest-mtime to resolve split-brain")
	selfHealSplitBrainCmd.Flags().StringVar(&flagSplitBrainSourceBrick, "source-brick", "", "Use a brick as source to resolve split-brain")
	selfHealSplitBrainCmd.Flags().StringVar(&flagFileName, "file", "", "Specify filename that is in split-brain")
	selfHealCmd.AddCommand(selfHealSplitBrainCmd)

	volumeCmd.AddCommand(selfHealCmd)
}

var selfHealInfoCmd = &cobra.Command{
	Use:   "info <volname> [--info-summary|--split-brain-info]",
	Short: "Self Heal Info",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var selfHealInfo []glustershdapi.BrickHealInfo
		volname := args[0]
		if flagSummaryInfo {
			selfHealInfo, err = client.SelfHealInfo(volname, "info-summary")
		} else if flagSplitBrainInfo {
			selfHealInfo, err = client.SelfHealInfo(volname, "split-brain-info")
		} else {
			selfHealInfo, err = client.SelfHealInfo(volname)
		}
		if err != nil {
			if GlobalFlag.Verbose {
				log.WithError(err).WithField("volume", volname).Error("failed to get heal info")
			}
			failure(fmt.Sprintf("Failed to get heal info for volume %s\n", volname), err, 1)
		}
		for index := range selfHealInfo {
			fmt.Printf("Brick: %s\n", selfHealInfo[index].Name)
			fmt.Printf("Status: %s\n", selfHealInfo[index].Status)

			if selfHealInfo[index].TotalEntries != nil {
				fmt.Printf("total-entries: %v\n", *selfHealInfo[index].TotalEntries)
			}
			if selfHealInfo[index].EntriesInHealPending != nil {
				fmt.Printf("entries-in-heal-pending: %v\n", *selfHealInfo[index].EntriesInHealPending)
			}
			if selfHealInfo[index].EntriesInSplitBrain != nil {
				fmt.Printf("entries-in-split-brain: %v\n", *selfHealInfo[index].EntriesInSplitBrain)
			}
			if selfHealInfo[index].EntriesPossiblyHealing != nil {
				fmt.Printf("entries-possibly-healing: %v\n", *selfHealInfo[index].EntriesPossiblyHealing)
			}
			if selfHealInfo[index].Entries != nil {
				fmt.Printf("entries: %v\n", *selfHealInfo[index].Entries)
			}
			if selfHealInfo[index].Files != nil {
				for value := range selfHealInfo[index].Files {
					fmt.Printf("%s:%s\n", selfHealInfo[index].Files[value].GfID, selfHealInfo[index].Files[value].Filename)
				}
			}
			fmt.Printf("\n")
		}
	},
}

var selfHealIndexCmd = &cobra.Command{
	Use:   "index <volname>",
	Short: "Index Heal",
	Long:  "CLI command to trigger index heal on a volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		volname := args[0]
		err = client.SelfHeal(volname, "index")
		if err != nil {
			failure(fmt.Sprintf("Failed to run heal for volume %s\n", volname), err, 1)
		}
		fmt.Println("Heal on volume has been successfully launched. Use heal info to check status")
	},
}

var selfHealFullCmd = &cobra.Command{
	Use:   "full <volname>",
	Short: "Full Heal",
	Long:  "CLI command to trigger full heal on a volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		volname := args[0]
		err = client.SelfHeal(volname, "full")
		if err != nil {
			failure(fmt.Sprintf("Failed to run heal for volume %s\n", volname), err, 1)
		}
		fmt.Println("Heal on volume has been successfully launched. Use heal info to check status")
	},
}

var selfHealSplitBrainCmd = &cobra.Command{
	Use:   "split-brain <volname> [--bigger-file --file <filename>|--latest-mtime --file <filename>|--source-brick <hostname:brickname> [--file <filename>]]",
	Short: "Split-brain operations",
	Long:  "Resolve split-brain situation based on bigger-file, latest-mtime or a source-brick",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var operation string
		var req glustershdapi.SplitBrainReq
		volname := args[0]
		if flagSplitBrainBiggerFile {
			if flagFileName == "" {
				failure("Split brain operation failed", errors.New("Please provide filename"), 1)
			}
			req.FileName = flagFileName
			operation = "bigger-file"
		} else if flagSplitBrainLatestMtime {
			if flagFileName == "" {
				failure("Split brain operation failed", errors.New("Please provide filename"), 1)
			}
			req.FileName = flagFileName
			operation = "latest-mtime"
		} else if flagSplitBrainSourceBrick != "" {
			hnameAndBrick := strings.Split(flagSplitBrainSourceBrick, ":")
			if len(hnameAndBrick) < 2 {
				failure("Split brain operation failed", errors.New("Please provide both hostname and brickpath"), 1)
			}
			req.HostName, req.BrickName = hnameAndBrick[0], hnameAndBrick[1]
			if flagFileName != "" {
				req.FileName = flagFileName
			}
			operation = "source-brick"
		} else {
			failure("Split brain operation failed", errors.New("Please provide a valid split-brain resolution operation"), 1)
		}
		err := client.SelfHealSplitBrain(volname, operation, req)
		if err != nil {
			failure(fmt.Sprintf("Failed to resolve split-brain for volume %s\n", volname), err, 1)
		}
		fmt.Printf("Split Brain Resolution successful on volume %s \n", volname)
	},
}
