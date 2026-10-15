package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker commands",
	Args:  cobra.NoArgs,
	RunE:  runDockerOverview,
}

var dockerList bool
var dockerTop bool
var dockerOverviewJSON bool
var dockerOverviewAll bool
var dockerOverviewLimit = 10

func runDockerOverview(cmd *cobra.Command, args []string) error {
	if dockerList && dockerTop {
		return fmt.Errorf("docker: choose either --list or --top")
	}
	if !dockerTop && (dockerOverviewAll || dockerOverviewLimit != 10) {
		return fmt.Errorf("docker: --all and --limit require --top")
	}

	switch {
	case dockerList:
		dockerLsJSON = dockerOverviewJSON
		return runDockerLs(cmd, args)
	case dockerTop:
		dockerTopJSON = dockerOverviewJSON
		dockerTopAll = dockerOverviewAll
		dockerTopLimit = dockerOverviewLimit
		return runDockerTop(cmd, args)
	default:
		dockerSummaryJSON = dockerOverviewJSON
		return runDockerSummary(cmd, args)
	}
}

func init() {
	dockerLsCmd.Hidden = true
	dockerSummaryCmd.Hidden = true
	dockerTopCmd.Hidden = true
	dockerCmd.AddCommand(dockerLsCmd)
	dockerCmd.AddCommand(dockerSummaryCmd)
	dockerCmd.AddCommand(dockerTopCmd)
	dockerCmd.Flags().BoolVar(&dockerList, "list", false, "List running containers")
	dockerCmd.Flags().BoolVar(&dockerTop, "top", false, "Show largest images and volumes")
	dockerCmd.Flags().BoolVar(&dockerOverviewAll, "all", false, "Include anonymous volumes with --top")
	dockerCmd.Flags().IntVar(&dockerOverviewLimit, "limit", 10, "Number of items to show with --top")
	dockerCmd.Flags().BoolVar(&dockerOverviewJSON, "json", false, "Output JSON")
}
