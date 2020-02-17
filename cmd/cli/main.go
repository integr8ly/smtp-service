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
	"fmt"
	"os"

	"github.com/integr8ly/smtp-service/pkg/sendgrid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

const (
	defaultOutputSecretName = "redhat-rhmi-smtp"
	exitCodeErrKnown        = 1
	exitCodeErrUnknown      = 2
)

var flagDebug = false
var logger = logrus.NewEntry(&logrus.Logger{
	Out:          os.Stderr,
	Formatter:    &logrus.TextFormatter{},
	ReportCaller: false,
	Level:        logrus.FatalLevel,
})

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cli [sub command]",
	Short: "commands for managing rhmi cluster api keys",
}

func exitSuccess(message string) {
	fmt.Fprintf(os.Stdout, message)
	os.Exit(0)
}

func exitError(message string, code int) {
	fmt.Fprintf(os.Stderr, message)
	os.Exit(code)
}

func setupSMTPDetailsClient(logger *logrus.Entry) (*sendgrid.Client, error) {
	smtpdetailsClient, err := sendgrid.NewDefaultClient(logger)
	if err != nil {
		logger.Fatalf("failed to create sendgrid details client: %v", err)
		return nil, errors.Wrap(err, "failed to setup sendgrid smtp details client")
	}
	return smtpdetailsClient, nil
}

func init() {
	cobra.OnInitialize(func() {
		if flagDebug {
			logger.Logger.SetLevel(logrus.DebugLevel)
		}
	})
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.main.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug output to stderr")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
