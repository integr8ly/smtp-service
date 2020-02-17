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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// refreshCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// refreshCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	refreshCmd.Flags().StringP("secret-name", "s", defaultOutputSecretName, "Name of the output secret")
}
