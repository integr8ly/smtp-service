package main

import (
	"fmt"

	"github.com/integr8ly/smtp-service/pkg/smtpdetails"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [cluster id]",
	Short: "get sendgrid api key id associated with [cluster id]",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		smtpDetailsClient, err := setupSMTPDetailsClient(logger)
		if err != nil {
			exitError("failed to setup smtp details client", exitCodeErrUnknown)
		}
		smtpDetails, err := smtpDetailsClient.Get(args[0])
		if err != nil {
			if smtpdetails.IsNotExistError(err) {
				exitError(fmt.Sprintf("api key for cluster %s not found", args[0]), exitCodeErrKnown)
			}
			exitError(fmt.Sprintf("unknown error: %v", err), exitCodeErrUnknown)
		}
		exitSuccess(smtpDetails.ID)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
