/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	createCmd.Flags().StringP("secret-name", "s", defaultOutputSecretName, "Name of the output secret")
}
