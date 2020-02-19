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
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug output to stderr")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
