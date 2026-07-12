package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "devsentry",
	Short:         "Devsentry CLI",
	Long:          "Devsentry is a command-line interface tool for managing and monitoring development environments.",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run:           runRootCmd,
}

func runRootCmd(cmd *cobra.Command, args []string) {
	fmt.Fprintln(cmd.OutOrStdout(), "Welcome to devsentry! Use 'devsentry --help' to see available commands.")
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(dockerCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(cleanCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
