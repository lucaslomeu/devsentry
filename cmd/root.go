package cmd

import (
	"fmt"
	"strings"

	"devsentry/internal/clierr"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "devsentry",
	Short:         "Devsentry CLI",
	Long:          "Devsentry is a command-line interface tool for managing and monitoring development environments.",
	Version:       Version,
	Args:          func(cmd *cobra.Command, args []string) error { return usageArgs("devsentry", args) },
	SilenceErrors: true,
	SilenceUsage:  true,
	Run:           runRootCmd,
}

var quiet bool

func runRootCmd(cmd *cobra.Command, args []string) {
	if quiet {
		return
	}
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
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return clierr.Wrap(clierr.ErrUsage, err)
	})
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress successful human-readable output")
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
	err := rootCmd.Execute()
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "unknown command") {
		return clierr.Wrap(clierr.ErrUsage, err)
	}
	return err
}
