package cmd

import (
	"fmt"

	"github.com/gluster/glusterd2/pkg/api"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	labelCreateHelpShort = "Create a label"
	labelCreateHelpLong  = "Create a label that can use to tag different objects. Label values will be created with default values if choose to omit, specific label values should provide using relevant flags."
)

var (
	flagSnapMaxHardLimit uint64
	flagSnapMaxSoftLimit uint64
	flagActivateOnCreate bool
	flagAutoDelete       bool
	flagDescription      string

	labelCreateCmd = &cobra.Command{
		Use:   "create <labelname>",
		Short: labelCreateHelpShort,
		Long:  labelCreateHelpLong,
		Args:  cobra.MinimumNArgs(1),
		Run:   labelCreateCmdRun,
	}
)

func init() {
	labelCreateCmd.Flags().Uint64Var(&flagSnapMaxHardLimit, "snap-max-hard-limit", 256, "Snapshot maximum hard limit count")
	labelCreateCmd.Flags().Uint64Var(&flagSnapMaxSoftLimit, "snap-max-soft-limit", 230, "Snapshot maximum soft limit count")
	labelCreateCmd.Flags().BoolVar(&flagActivateOnCreate, "activate-on-create", false, "If enabled, Further snapshots will be activated after creation")
	labelCreateCmd.Flags().BoolVar(&flagAutoDelete, "auto-delete", false, "If enabled, Snapshots will be deleted upon reaching snap-max-soft-limit. If disabled A warning log will be generated")
	labelCreateCmd.Flags().StringVar(&flagDescription, "description", "", "Label description")

	labelCmd.AddCommand(labelCreateCmd)
}

func labelCreateCmdRun(cmd *cobra.Command, args []string) {
	labelname := args[0]

	req := api.LabelCreateReq{
		Name:             labelname,
		SnapMaxHardLimit: flagSnapMaxHardLimit,
		SnapMaxSoftLimit: flagSnapMaxSoftLimit,
		ActivateOnCreate: flagActivateOnCreate,
		AutoDelete:       flagAutoDelete,
		Description:      flagDescription,
	}

	info, err := client.LabelCreate(req)
	if err != nil {
		if GlobalFlag.Verbose {
			log.WithError(err).WithFields(
				log.Fields{
					"labelname": labelname,
				}).Error("label creation failed")
		}
		failure("Label creation failed", err, 1)
	}
	fmt.Printf("%s Label created successfully\n", info.Name)
}
