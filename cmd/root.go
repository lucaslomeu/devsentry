package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "devsentry",
	Short:         "Devsentry CLI",
	Long:          "Devsentry is a command-line interface tool for managing and monitoring development environments.",
	Version:       Version,
	Args:          cobra.NoArgs,
	SilenceErrors: true,
	SilenceUsage:  true,
	Run:           runRootCmd,
}

func runRootCmd(cmd *cobra.Command, args []string) {
	out := cmd.OutOrStdout()
	style := newOutputStyle(out)
	fmt.Fprintln(out, style.title("Devsentry"))
	fmt.Fprintln(out, style.muted("Development environment toolkit"))
	fmt.Fprintln(out)
	fmt.Fprintln(out, style.section("Quick actions"))
	fmt.Fprintf(out, "  %s  %s\n", style.info("devsentry docker"), "Docker overview")
	fmt.Fprintf(out, "  %s  %s\n", style.info("devsentry clean"), "Preview safe cleanup")
	fmt.Fprintf(out, "  %s  %s\n", style.info("devsentry status"), "Inspect macOS status")
	fmt.Fprintln(out)
	fmt.Fprintln(out, style.muted("Use 'devsentry --help' for all commands."))
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetVersionTemplate("devsentry version: {{.Version}}\n")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(dockerCmd)
	statusCmd.Hidden = true
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(lsAliasCmd)
	rootCmd.AddCommand(summaryAliasCmd)
	rootCmd.AddCommand(topAliasCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
