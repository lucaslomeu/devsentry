package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current devsentry CLI version.
// Overridden at build time via -ldflags.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the devsentry version",
	Run:   runVersionCmd,
}

func runVersionCmd(cmd *cobra.Command, args []string) {
	fmt.Fprintln(cmd.OutOrStdout(), "devsentry version:", Version)
}