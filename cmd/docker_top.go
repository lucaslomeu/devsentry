package cmd

import (
	"fmt"
	"text/tabwriter"

	"devsentry/internal/docker"

	"github.com/spf13/cobra"
)

var dockerTopAll bool
var dockerTopLimit = 10
var dockerTopJSON bool

var dockerTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Show the largest Docker images and volumes",
	Args:  func(cmd *cobra.Command, args []string) error { return usageArgs("docker top", args) },
	RunE:  runDockerTop,
}

var topAliasCmd = &cobra.Command{
	Use:    "top",
	Short:  "Show the largest Docker images and volumes",
	Hidden: true,
	Args:   func(cmd *cobra.Command, args []string) error { return usageArgs("top", args) },
	RunE:   runDockerTop,
}

func runDockerTop(cmd *cobra.Command, args []string) error {
	if dockerTopLimit <= 0 {
		return usageError("docker top: --limit must be greater than zero")
	}

	result, err := docker.GetTop(cmd.Context(), docker.TopOptions{
		IncludeAnonymousVolumes: dockerTopAll,
		Limit:                   dockerTopLimit,
		ImageFilters:            dockerFilterImages,
		VolumeFilters:           dockerFilterVolumes,
	})
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if dockerTopJSON {
		return writeCommandJSON(out, "docker.top", result)
	}
	if quiet {
		return nil
	}

	style := newOutputStyle(out)
	fmt.Fprintln(out, style.title("Devsentry - Docker Top"))
	fmt.Fprintln(out)

	fmt.Fprintf(out, "%s\n", style.section(fmt.Sprintf("Top %d Images by Size", dockerTopLimit)))
	if len(result.Images) == 0 {
		fmt.Fprintln(out, "  None")
	} else {
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		for i, item := range result.Images {
			fmt.Fprintf(w, "  %d.\t%s\t%s\n", i+1, item.Name, style.info(item.SizeLabel))
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	fmt.Fprintln(out)

	fmt.Fprintf(out, "%s\n", style.section(fmt.Sprintf("Top %d Volumes by Size", dockerTopLimit)))
	if len(result.Volumes) == 0 {
		fmt.Fprintln(out, "  None")
	} else {
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		for i, item := range result.Volumes {
			fmt.Fprintf(w, "  %d.\t%s\t%s\n", i+1, item.Name, style.info(item.SizeLabel))
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	for _, command := range []*cobra.Command{dockerTopCmd, topAliasCmd} {
		command.Flags().BoolVar(&dockerTopAll, "all", false, "Include anonymous volumes")
		command.Flags().IntVar(&dockerTopLimit, "limit", 10, "Number of images and volumes to show")
		command.Flags().BoolVar(&dockerTopJSON, "json", false, "Output JSON")
		command.Flags().StringSliceVar(&dockerFilterImages, "image", nil, "Filter images by reference")
		command.Flags().StringSliceVar(&dockerFilterVolumes, "volume", nil, "Filter volumes by name")
	}
}
