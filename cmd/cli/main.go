package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/integr8ly/smtp-service/version"

	"github.com/integr8ly/smtp-service/pkg/smtpdetails"

	"github.com/integr8ly/smtp-service/pkg/sendgrid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	defaultOutputSecretName = "rhmi-smtp"
	exitCodeErrKnown        = 1
	exitCodeErrUnknown      = 2
)

var (
	flagDebug string
)

func init() {
	pflag.StringVarP(&flagDebug, "debug", "d", "error", "--debug=[info|verbose|error]")
}

func main() {
	pflag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(getDebugLevelFromString(flagDebug))
	logrus.SetOutput(os.Stderr)
	logger := logrus.WithField("service", "rhmi_sendgrid_cli")
	smtpdetailsClient, err := sendgrid.NewDefaultClient(logger)
	if err != nil {
		logger.Fatalf("failed to create sendgrid details client: %v", err)
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
				if smtpdetails.IsAlreadyExistsError(err) {
					exitError(fmt.Sprintf("api key for cluster %s already exists", args[0]), exitCodeErrKnown)
				}
				exitError(fmt.Sprintf("unknown error: %v", err), exitCodeErrUnknown)
			}
			logger.Debug("smtp details created successfully, converting to secret")
			smtpSecret := smtpdetails.ConvertSMTPDetailsToSecret(smtpDetails, defaultOutputSecretName)
			smtpJSON, err := json.MarshalIndent(smtpSecret, "", "    ")
			if err != nil {
				exitError(fmt.Sprintf("error converting details to secret: %v", err), exitCodeErrUnknown)
			}
			exitSuccess(string(smtpJSON))
		},
	}

	cmdGet := &cobra.Command{
		Use:   "get [cluster id]",
		Short: "get sendgrid api key id associated with [cluster id]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			smtpDetails, err := smtpdetailsClient.Get(args[0])
			if err != nil {
				if smtpdetails.IsNotExistError(err) {
					exitError(fmt.Sprintf("api key for cluster %s not found", args[0]), exitCodeErrKnown)
				}
				exitError(fmt.Sprintf("unknown error: %v", err), exitCodeErrUnknown)
			}
			exitSuccess(smtpDetails.ID)
		},
	}

	cmdDelete := &cobra.Command{
		Use:   "delete [cluster id]",
		Short: "delete sendgrid sub user and api key associated with [cluster id]",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := smtpdetailsClient.Delete(args[0]); err != nil {
				if smtpdetails.IsNotExistError(err) {
					exitError(fmt.Sprintf("api key for cluster %s does not exist", args[0]), exitCodeErrKnown)
				}
				exitError(fmt.Sprintf("failed to delete api key %v", err), exitCodeErrUnknown)
			}
			exitSuccess("api key deleted")
		},
	}

	cmdRefresh := &cobra.Command{
		Use:   "refresh [cluster id]",
		Short: "delete api key associated with [cluster id] and genereate a new key",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key, err := smtpdetailsClient.Refresh(args[0])
			if err != nil {
				if smtpdetails.IsNotExistError(err) {
					exitError(fmt.Sprintf("api key for cluster %s does not exist", args[0]), exitCodeErrKnown)
				}
				exitError(fmt.Sprintf("failed to delete api key %v", err), exitCodeErrUnknown)
			}
			exitSuccess(key)
		},
	}

	cmdVersion := &cobra.Command{
		Use:   "version",
		Short: "print the version of the cli",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(version.Version)
		},
	}

	rootCmd.AddCommand(cmdDelete, cmdGet, cmdCreate, cmdRefresh, cmdVersion)
	if err = rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func exitSuccess(message string) {
	fmt.Fprintf(os.Stdout, message)
	os.Exit(0)
}

func exitError(message string, code int) {
	fmt.Fprintf(os.Stderr, message)
	os.Exit(code)
}

func getDebugLevelFromString(levelStr string) logrus.Level {
	debugMap := map[string]logrus.Level{
		"verbose": logrus.DebugLevel,
		"info":    logrus.InfoLevel,
	}
	level := debugMap[levelStr]
	if level == 0 {
		return logrus.ErrorLevel
	}
	return level
}
