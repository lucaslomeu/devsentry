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
	Args:  cobra.NoArgs,
	RunE:  runDockerLs,
}

var lsAliasCmd = &cobra.Command{
	Use:    "ls",
	Short:  "List running Docker containers",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE:   runDockerLs,
}

var dockerLsJSON bool

func runDockerLs(cmd *cobra.Command, args []string) error {
	containers, err := docker.ListRunningContainers(cmd.Context())
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if dockerLsJSON {
		return writeJSON(out, containers)
	}

	style := newOutputStyle(out)
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, style.header("NAME\tIMAGE\tSTATUS\tPORTS"))
	for _, c := range containers {
		ports := c.Ports
		if ports == "" {
			ports = style.muted("-")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name, c.Image, style.state(c.Status), ports)
	}

	return w.Flush()
}

func init() {
	dockerLsCmd.Flags().BoolVar(&dockerLsJSON, "json", false, "Output JSON")
	lsAliasCmd.Flags().BoolVar(&dockerLsJSON, "json", false, "Output JSON")
}
