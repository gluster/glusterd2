package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	helpEventsCmd              = "Gluster Events"
	helpEventsWebhookAddCmd    = ""
	helpEventsWebhookDeleteCmd = ""
	helpEventsWebhookListCmd   = ""
)

var (
	// Create Command Flags
	flagWebhookAddCmdToken  string
	flagWebhookAddCmdSecret string
)

func init() {
	eventsWebhookAddCmd.Flags().StringVarP(&flagWebhookAddCmdToken, "bearer-token", "t", "", "Bearer Token")
	eventsWebhookAddCmd.Flags().StringVarP(&flagWebhookAddCmdSecret, "secret", "s", "", "Secret to generate JWT Bearer Token")

	eventsCmd.AddCommand(eventsWebhookAddCmd)

	eventsCmd.AddCommand(eventsWebhookDeleteCmd)

	eventsCmd.AddCommand(eventsWebhookListCmd)

	RootCmd.AddCommand(eventsCmd)
}

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: helpEventsCmd,
}

var eventsWebhookAddCmd = &cobra.Command{
	Use:   "webhook-add [flags] <URL>",
	Short: helpEventsWebhookAddCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		err := client.WebhookAdd(url, flagWebhookAddCmdToken, flagWebhookAddCmdSecret)
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"url":   url,
					"error": err.Error(),
				}).Error("failed to add webhook")
			}
			failure("Failed to add Webhook", err, 1)
		}
		fmt.Printf("Webhook %s added successfully\n", url)
	},
}

var eventsWebhookDeleteCmd = &cobra.Command{
	Use:   "webhook-del <URL>",
	Short: helpEventsWebhookDeleteCmd,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		err := client.WebhookDelete(url)
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"url":   url,
					"error": err.Error(),
				}).Error("failed to delete webhook")
			}
			failure("Failed to delete Webhook", err, 1)
		}
		fmt.Printf("Webhook %s deleted successfully\n", url)
	},
}

var eventsWebhookListCmd = &cobra.Command{
	Use:   "webhooks",
	Short: helpEventsWebhookAddCmd,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		webhooks, err := client.Webhooks()
		if err != nil {
			if verbose {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("failed to get list of Webhooks")
			}
			failure("Failed to get list of registered Webhooks", err, 1)
		}

		if len(webhooks) > 0 {
			fmt.Printf("Webhooks:\n%s\n", strings.Join(webhooks, "\n"))
		}
	},
}
