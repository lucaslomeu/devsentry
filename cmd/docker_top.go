package cmd

import (
	"fmt"
	"text/tabwriter"

	"devsentry/internal/docker"

	"github.com/spf13/cobra"
)

var dockerTopAll bool

var dockerTopCmd = &cobra.Command{
	Use:   "top",
	Short: "Show the largest Docker images and volumes",
	RunE:  runDockerTop,
}

func runDockerTop(cmd *cobra.Command, args []string) error {
	result, err := docker.GetTop(cmd.Context(), docker.TopOptions{
		IncludeAnonymousVolumes: dockerTopAll,
	})
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Devsentry - Docker Top")
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Top 10 Images by Size:")
	if len(result.Images) == 0 {
		fmt.Fprintln(out, "  None")
	} else {
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		for i, item := range result.Images {
			fmt.Fprintf(w, "  %d.\t%s\t%s\n", i+1, item.Name, item.SizeLabel)
		}
		_ = w.Flush()
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Top 10 Volumes by Size:")
	if len(result.Volumes) == 0 {
		fmt.Fprintln(out, "  None")
	} else {
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		for i, item := range result.Volumes {
			fmt.Fprintf(w, "  %d.\t%s\t%s\n", i+1, item.Name, item.SizeLabel)
		}
		_ = w.Flush()
	}

	return nil
}
