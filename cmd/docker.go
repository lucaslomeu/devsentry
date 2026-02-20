package cmd

import "github.com/spf13/cobra"

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker commands",
}

func init() {
	dockerCmd.AddCommand(dockerLsCmd)
	dockerCmd.AddCommand(dockerSummaryCmd)
}
