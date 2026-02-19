package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devsentry",
	Short: "CLI do devsentry",
	Long:  "Devsentry is a command-line interface tool for managing and monitoring development environments.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "Bem-vindo ao devsentry!")
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
