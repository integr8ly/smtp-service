package main

import (
	"encoding/json"
	"fmt"

	"github.com/integr8ly/smtp-service/pkg/smtpdetails"
	"github.com/spf13/cobra"
)

// refreshCmd represents the refresh command
var refreshCmd = &cobra.Command{
	Use:   "refresh [cluster id]",
	Short: "delete api key associated with [cluster id] and genereate a new key",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		smtpDetailsClient, err := setupSMTPDetailsClient(logger)
		if err != nil {
			exitError("failed to setup smtp details client", exitCodeErrUnknown)
		}
		smtpDetails, err := smtpDetailsClient.Refresh(args[0])
		if err != nil {
			if smtpdetails.IsNotExistError(err) {
				exitError(fmt.Sprintf("cannot create api key for cluster that does not exist, cluster=%s, use the create command", args[0]), exitCodeErrKnown)
			}
			exitError(fmt.Sprintf("failed to delete api key %v: ", err), exitCodeErrUnknown)
		}
		secretName, err := cmd.Flags().GetString("secret-name")
		if err != nil {
			exitError("failed to get secret name flag", exitCodeErrUnknown)
		}
		if secretName == "" {
			logger.Infof("secret name is blank, using default name %s", defaultOutputSecretName)
			secretName = defaultOutputSecretName
		}
		smtpSecret := smtpdetails.ConvertSMTPDetailsToSecret(smtpDetails, secretName)
		smtpJSON, err := json.MarshalIndent(smtpSecret, "", "    ")
		if err != nil {
			exitError(fmt.Sprintf("error converting details to secret: %v", err), exitCodeErrUnknown)
		}
		exitSuccess(string(smtpJSON))
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)
	refreshCmd.Flags().StringP("secret-name", "s", defaultOutputSecretName, "Name of the output secret")
}
