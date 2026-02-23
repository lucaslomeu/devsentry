package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devsentry",
	Short: "Devsentry CLI",
	Long:  "Devsentry is a command-line interface tool for managing and monitoring development environments.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "Welcome to devsentry!")
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(dockerCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(summaryCmd)
	rootCmd.AddCommand(topCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(cleanCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
