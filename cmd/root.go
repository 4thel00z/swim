package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "swim",
	Short: "A CLI tool to manage Docker containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return updatePortCmd.RunE(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {}
