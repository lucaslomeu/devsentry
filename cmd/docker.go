package cmd

import "github.com/spf13/cobra"

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker commands",
	Args:  func(cmd *cobra.Command, args []string) error { return usageArgs("docker", args) },
	RunE:  runDockerOverview,
}

var dockerList bool
var dockerTop bool
var dockerOverviewJSON bool
var dockerOverviewAll bool
var dockerOverviewLimit = 10

var (
	dockerFilterNames    []string
	dockerFilterImages   []string
	dockerFilterStatuses []string
	dockerFilterVolumes  []string
)

func runDockerOverview(cmd *cobra.Command, args []string) error {
	if dockerList && dockerTop {
		return usageError("docker: choose either --list or --top")
	}
	if !dockerTop && (dockerOverviewAll || dockerOverviewLimit != 10) {
		return usageError("docker: --all and --limit require --top")
	}
	if len(dockerFilterStatuses) > 0 && !dockerList {
		return usageError("docker: --status requires --list")
	}
	if len(dockerFilterNames) > 0 && !dockerList {
		return usageError("docker: --name requires --list")
	}
	if len(dockerFilterVolumes) > 0 && !dockerTop {
		return usageError("docker: --volume requires --top")
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
	dockerCmd.Flags().StringSliceVar(&dockerFilterNames, "name", nil, "Filter containers by name with --list")
	dockerCmd.Flags().StringSliceVar(&dockerFilterImages, "image", nil, "Filter by image")
	dockerCmd.Flags().StringSliceVar(&dockerFilterStatuses, "status", nil, "Filter containers by status with --list")
	dockerCmd.Flags().StringSliceVar(&dockerFilterVolumes, "volume", nil, "Filter volumes by name with --top")
}
