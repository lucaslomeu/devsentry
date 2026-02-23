package cmd

import (
	"fmt"
	"text/tabwriter"

	"devsentry/internal/docker"

	"github.com/spf13/cobra"
)

var dockerLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List running Docker containers",
	RunE:  runDockerLs,
}

func runDockerLs(cmd *cobra.Command, args []string) error {
	containers, err := docker.ListRunningContainers(cmd.Context())
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tIMAGE\tSTATUS\tPORTS")
	for _, c := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name, c.Image, c.Status, c.Ports)
	}

	return w.Flush()
}
