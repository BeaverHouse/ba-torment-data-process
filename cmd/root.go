package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Values arrive in the environment, put there by whoever starts the process:
// `austincli runbook task run` locally and the container in production. Loading
// a .env file was how this used to work, and it made the CLI unrunnable on a
// machine that had never generated one.
var rootCmd = &cobra.Command{
	Use:   "batorment",
	Short: "Blue Archive torment data processing CLI",
	Long:  `CLI tool for processing and managing Blue Archive torment raid data (parties, analysis, grid images, etc.).`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
