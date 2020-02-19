package main

import (
	"encoding/json"
	"fmt"

	"github.com/integr8ly/smtp-service/pkg/smtpdetails"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [cluster id]",
	Short: "create sendgrid sub user and api key associated with [cluster id]",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		smtpDetailsClient, err := setupSMTPDetailsClient(logger)
		if err != nil {
			exitError("failed to setup smtp details client", exitCodeErrUnknown)
		}
		smtpDetails, err := smtpDetailsClient.Create(args[0])
		if err != nil {
			if smtpdetails.IsAlreadyExistsError(err) {
				exitError(fmt.Sprintf("api key for cluster %s already exists", args[0]), exitCodeErrKnown)
			}
			exitError(fmt.Sprintf("unknown error: %v", err), exitCodeErrUnknown)
		}
		logger.Debug("smtp details created successfully, converting to secret")
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
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("secret-name", "s", defaultOutputSecretName, "Name of the output secret")
}
