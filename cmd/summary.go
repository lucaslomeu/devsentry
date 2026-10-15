package cmd

import (
	"fmt"
	"strings"

	"devsentry/internal/docker"

	"github.com/spf13/cobra"
)

var dockerSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show a summary of Docker resources",
	Args:  func(cmd *cobra.Command, args []string) error { return usageArgs("docker summary", args) },
	RunE:  runDockerSummary,
}

var summaryAliasCmd = &cobra.Command{
	Use:    "summary",
	Short:  "Show a summary of Docker resources",
	Hidden: true,
	Args:   func(cmd *cobra.Command, args []string) error { return usageArgs("summary", args) },
	RunE:   runDockerSummary,
}

var dockerSummaryJSON bool

func runDockerSummary(cmd *cobra.Command, args []string) error {
	summary, err := docker.GetSummary(cmd.Context())
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if dockerSummaryJSON {
		return writeCommandJSON(out, "docker", summary)
	}
	if quiet {
		return nil
	}

	style := newOutputStyle(out)
	fmt.Fprintln(out, style.title("Devsentry - Docker Summary"))
	fmt.Fprintln(out, style.muted(strings.Repeat("-", 32)))
	fmt.Fprintln(out)

	fmt.Fprintln(out, style.section("Images"))
	fmt.Fprintf(out, "  Total: %s\n", style.info(fmt.Sprintf("%d", summary.Images)))
	fmt.Fprintf(out, "  Total Size: %s\n", style.info(normalizeSize(summary.ImagesSize)))
	fmt.Fprintln(out)

	fmt.Fprintln(out, style.section("Containers"))
	fmt.Fprintf(out, "  Running: %s\n", style.success(fmt.Sprintf("%d", summary.ContainersRunning)))
	stoppedStyle := style.info
	if summary.ContainersStopped > 0 {
		stoppedStyle = style.warning
	}
	fmt.Fprintf(out, "  Stopped: %s\n", stoppedStyle(fmt.Sprintf("%d", summary.ContainersStopped)))
	fmt.Fprintln(out)

	fmt.Fprintln(out, style.section("Volumes"))
	fmt.Fprintf(out, "  Total: %s\n", style.info(fmt.Sprintf("%d", summary.Volumes)))
	fmt.Fprintf(out, "  Total Size: %s\n", style.info(normalizeSize(summary.VolumesSize)))
	fmt.Fprintln(out)

	fmt.Fprintln(out, style.section("Build Cache"))
	fmt.Fprintf(out, "  Total: %s\n", style.info(fmt.Sprintf("%d", summary.BuildCache)))
	fmt.Fprintf(out, "  Total Size: %s\n", style.info(normalizeSize(summary.BuildCacheSize)))

	return nil
}

func init() {
	dockerSummaryCmd.Flags().BoolVar(&dockerSummaryJSON, "json", false, "Output JSON")
	summaryAliasCmd.Flags().BoolVar(&dockerSummaryJSON, "json", false, "Output JSON")
}

func normalizeSize(size string) string {
	if size == "" {
		return "0B"
	}
	return size
}
