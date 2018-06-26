package cmd

import (
	"fmt"

	glustershdapi "github.com/gluster/glusterd2/plugins/glustershd/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Heal Info Flags
	flagSummaryInfo    bool
	flagSplitBrainInfo bool
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
			if verbose {
				log.WithFields(log.Fields{
					"volume": volname,
					"error":  err.Error(),
				}).Error("failed to get heal info")
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
