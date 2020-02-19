package main

import (
	"github.com/integr8ly/smtp-service/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the version of the cli",
	Run: func(cmd *cobra.Command, args []string) {
		exitSuccess(version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
