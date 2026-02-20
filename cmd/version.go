package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current devsentry CLI version.
var Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versao do devsentry",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), Version)
	},
}
