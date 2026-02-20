package cmd

import (
	"context"
	"fmt"
	"text/tabwriter"

	"devsentry/internal/docker"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Lista containers Docker em execucao",
	RunE: func(cmd *cobra.Command, args []string) error {
		containers, err := docker.ListRunningContainers(context.Background())
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tIMAGE\tSTATUS\tPORTS")
		for _, c := range containers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name, c.Image, c.Status, c.Ports)
		}

		return w.Flush()
	},
}
