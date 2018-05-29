package cmd

import (
	"fmt"
	"os"

	"github.com/gluster/glusterd2/pkg/glustershd/api"

	"github.com/olekukonko/tablewriter"
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

	volumeCmd.AddCommand(selfHealCmd)
}

var selfHealInfoCmd = &cobra.Command{
	Use:   "info <volname> [--info-summary|--split-brain-info]",
	Short: "Self Heal Info",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var selfHealInfo []api.BrickHealInfo
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
		table := tablewriter.NewWriter(os.Stdout)
		var tableHeader []string
		for index := range selfHealInfo {
			var tableValues []string
			tableHeader = append(tableHeader, "Brick")
			tableValues = append(tableValues, *selfHealInfo[index].Name)
			tableHeader = append(tableHeader, "Status")
			tableValues = append(tableValues, *selfHealInfo[index].Status)
			if selfHealInfo[index].TotalEntries != nil {
				tableHeader = append(tableHeader, "total-entries")
				tableValues = append(tableValues, fmt.Sprintf("%v", *selfHealInfo[index].TotalEntries))
			}
			if selfHealInfo[index].EntriesInHealPending != nil {
				tableHeader = append(tableHeader, "entries-in-heal-pending")
				tableValues = append(tableValues, fmt.Sprintf("%v", *selfHealInfo[index].EntriesInHealPending))
			}
			if selfHealInfo[index].EntriesInSplitBrain != nil {
				tableHeader = append(tableHeader, "entries-in-split-brain")
				tableValues = append(tableValues, fmt.Sprintf("%v", *selfHealInfo[index].EntriesInSplitBrain))
			}
			if selfHealInfo[index].EntriesPossiblyHealing != nil {
				tableHeader = append(tableHeader, "entries-possibly-healing")
				tableValues = append(tableValues, fmt.Sprintf("%v", *selfHealInfo[index].EntriesPossiblyHealing))
			}
			if selfHealInfo[index].Entries != nil {
				tableHeader = append(tableHeader, "entries")
				tableValues = append(tableValues, fmt.Sprintf("%v", *selfHealInfo[index].Entries))
			}
			if index == 0 {
				table.SetHeader(tableHeader)
			}
			table.Append(tableValues)
		}
		table.Render()
	},
}
