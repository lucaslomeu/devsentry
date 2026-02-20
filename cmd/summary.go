package cmd

import (
	"fmt"

	"devsentry/internal/docker"

	"github.com/spf13/cobra"
)

var dockerSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show a summary of Docker resources",
	RunE:  runDockerSummary,
}

var summaryCmd = &cobra.Command{
	Use:    "summary",
	Short:  "Alias for devsentry docker summary",
	Hidden: true,
	RunE:   runDockerSummary,
}

func runDockerSummary(cmd *cobra.Command, args []string) error {
	summary, err := docker.GetSummary(cmd.Context())
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Devsentry - Docker Summary")
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Images:")
	fmt.Fprintf(out, "  Total: %d\n", summary.Images)
	fmt.Fprintf(out, "  Total Size: %s\n", normalizeSize(summary.ImagesSize))
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Containers:")
	fmt.Fprintf(out, "  Running: %d\n", summary.ContainersRunning)
	fmt.Fprintf(out, "  Stopped: %d\n", summary.ContainersStopped)
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Volumes:")
	fmt.Fprintf(out, "  Total: %d\n", summary.Volumes)
	fmt.Fprintf(out, "  Total Size: %s\n", normalizeSize(summary.VolumesSize))
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Build Cache:")
	fmt.Fprintf(out, "  Total: %d\n", summary.BuildCache)
	fmt.Fprintf(out, "  Total Size: %s\n", normalizeSize(summary.BuildCacheSize))

	return nil
}

func normalizeSize(size string) string {
	if size == "" {
		return "0B"
	}
	return size
}
