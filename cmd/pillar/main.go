// Package main is the entry point to the Mattermost Customer Web Server and CLI.
package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "pillar",
	Short: "Pillar is a service for customer support to view and edit Mattermost Cloud workspaces.",
	Run: func(cmd *cobra.Command, args []string) {
		_ = serverCmd.RunE(cmd, args)
	},
	// SilenceErrors allows us to explicitly log the error returned from rootCmd below.
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(workspaceCmd)
}

func main() {
	viper.SetEnvPrefix("PILLAR")
	viper.AutomaticEnv()
	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Error("command failed")
		os.Exit(1)
	}
}
