package main

import (
	"fmt"

	"github.com/integr8ly/smtp-service/pkg/smtpdetails"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete [cluster id]",
	Short: "delete sendgrid sub user and api key associated with [cluster id]",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		smtpDetailsClient, err := setupSMTPDetailsClient(logger)
		if err != nil {
			exitError("failed to setup smtp details client", exitCodeErrUnknown)
		}
		if err := smtpDetailsClient.Delete(args[0]); err != nil {
			if smtpdetails.IsNotExistError(err) {
				exitError(fmt.Sprintf("api key for cluster %s does not exist: %+v", args[0], err), exitCodeErrKnown)
			}
			exitError(fmt.Sprintf("failed to delete api key %v", err), exitCodeErrUnknown)
		}
		exitSuccess("api key deleted")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
