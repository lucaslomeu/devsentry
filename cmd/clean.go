package cmd

import (
	"fmt"

	"devsentry/internal/clean"

	"github.com/spf13/cobra"
)

var cleanApply bool

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up Docker resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		plan, err := clean.BuildPlan(cmd.Context())
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		mode := "DRY-RUN"
		if cleanApply {
			mode = "APPLY"
		}

		fmt.Fprintf(out, "Devsentry - Docker Clean (%s)\n\n", mode)

		fmt.Fprintln(out, "Dangling Images:")
		if len(plan.DanglingImages) == 0 {
			fmt.Fprintln(out, "  None")
		} else {
			for _, img := range plan.DanglingImages {
				fmt.Fprintf(out, "  %s\n", img.ID)
			}
		}
		fmt.Fprintln(out)

		fmt.Fprintln(out, "Stopped Containers:")
		if len(plan.StoppedContainers) == 0 {
			fmt.Fprintln(out, "  None")
		} else {
			for _, ctr := range plan.StoppedContainers {
				name := ctr.Name
				if name == "" {
					name = "<unnamed>"
				}
				fmt.Fprintf(out, "  container %s (%s)\n", ctr.ID, name)
			}
		}
		fmt.Fprintln(out)

		fmt.Fprintln(out, "Build Cache:")
		if plan.BuildCache.Available {
			suffix := ""
			if plan.BuildCache.Approx {
				suffix = " (estimated)"
			}
			fmt.Fprintf(out, "  %s reclaimable%s\n", plan.BuildCache.SizeLabel, suffix)
		} else {
			fmt.Fprintln(out, "  Unknown")
		}
		fmt.Fprintln(out)

		fmt.Fprintln(out, "Planned Actions:")
		for _, action := range plan.PlannedActions {
			fmt.Fprintf(out, "  %s\n", action)
		}
		fmt.Fprintln(out)

		if plan.EstimatedBytes > 0 {
			prefix := ""
			if plan.EstimateApprox {
				prefix = "~"
			}
			fmt.Fprintf(out, "Estimated Space to Reclaim:\n  %s%s\n\n", prefix, formatBytes(plan.EstimatedBytes))
		} else {
			fmt.Fprintln(out, "Estimated Space to Reclaim:\n  Unknown\n")
		}

		if !cleanApply {
			fmt.Fprintln(out, "Nothing executed (dry-run mode).")
			return nil
		}

		if err := clean.Apply(cmd.Context()); err != nil {
			return err
		}

		fmt.Fprintln(out, "Cleanup executed.")
		return nil
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanApply, "apply", false, "Execute the cleanup commands")
}

func formatBytes(bytes uint64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
		tb = 1024 * gb
	)

	switch {
	case bytes >= tb:
		return fmt.Sprintf("%.1fTiB", float64(bytes)/float64(tb))
	case bytes >= gb:
		return fmt.Sprintf("%.1fGiB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1fMiB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1fKiB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
