package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/integr8ly/smtp-service/version"

	"github.com/integr8ly/smtp-service/pkg/smtpdetails"

	"github.com/integr8ly/smtp-service/pkg/sendgrid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	defaultOutputSecretName = "rhmi-smtp"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.ErrorLevel)
	logrus.SetOutput(os.Stderr)
	logger := logrus.WithField("service", "rhmi_sendgrid_cli")
	smtpdetailsClient, err := sendgrid.NewDefaultClient(logger)
	if err != nil {
		logger.Fatalf("failed to create sendgrid details client")
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "cli [sub command]",
		Short: "commands for managing rhmi cluster api keys",
	}

	cmdCreate := &cobra.Command{
		Use:   "create [cluster id]",
		Short: "create sendgrid sub user and api key associated with [cluster id]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			smtpDetails, err := smtpdetailsClient.Create(args[0])
			if err != nil {
				logger.Fatal("failed to create api key", err)
			}
			logger.Debug("smtp details created successfully, converting to secret")
			smtpSecret := smtpdetails.ConvertSMTPDetailsToSecret(smtpDetails, defaultOutputSecretName)
			smtpJSON, err := json.MarshalIndent(smtpSecret, "", "    ")
			if err != nil {
				logger.Fatal("failed to convert smtp secret to json", err)
			}
			fmt.Print(string(smtpJSON))
		},
	}

	cmdGet := &cobra.Command{
		Use:   "get [cluster id]",
		Short: "get sendgrid api key id associated with [cluster id]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			smtpDetails, err := smtpdetailsClient.Get(args[0])
			if err != nil {
				logger.Fatal("failed to get api key details", err)
			}
			fmt.Print(smtpDetails.ID)
		},
	}

	cmdDelete := &cobra.Command{
		Use:   "delete [cluster id]",
		Short: "delete sendgrid sub user and api key associated with [cluster id]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := smtpdetailsClient.Delete(args[0]); err != nil {
				logger.Fatal("failed to delete api key", err)
			}
		},
	}

	cmdVersion := &cobra.Command{
		Use:   "version",
		Short: "print the version of the cli",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(version.Version)
		},
	}

	rootCmd.AddCommand(cmdDelete, cmdGet, cmdCreate, cmdVersion)
	if err = rootCmd.Execute(); err != nil {
		panic(err)
	}
}
