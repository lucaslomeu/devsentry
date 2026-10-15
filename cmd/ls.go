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
	Args:  func(cmd *cobra.Command, args []string) error { return usageArgs("docker ls", args) },
	RunE:  runDockerLs,
}

var lsAliasCmd = &cobra.Command{
	Use:    "ls",
	Short:  "List running Docker containers",
	Hidden: true,
	Args:   func(cmd *cobra.Command, args []string) error { return usageArgs("ls", args) },
	RunE:   runDockerLs,
}

var dockerLsJSON bool

func runDockerLs(cmd *cobra.Command, args []string) error {
	containers, err := docker.ListRunningContainersWithFilters(cmd.Context(), docker.ContainerFilters{
		Names:    dockerFilterNames,
		Images:   dockerFilterImages,
		Statuses: dockerFilterStatuses,
	})
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if dockerLsJSON {
		return writeCommandJSON(out, "docker.list", containers)
	}
	if quiet {
		return nil
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
	for _, command := range []*cobra.Command{dockerLsCmd, lsAliasCmd} {
		command.Flags().StringSliceVar(&dockerFilterNames, "name", nil, "Filter containers by name")
		command.Flags().StringSliceVar(&dockerFilterImages, "image", nil, "Filter containers by image")
		command.Flags().StringSliceVar(&dockerFilterStatuses, "status", nil, "Filter containers by status")
	}
}
