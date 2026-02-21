package cmd

import (
	"fmt"

	"devsentry/internal/status"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Local environment status",
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := status.GetStatus(cmd.Context(), status.Options{})
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Devsentry - Environment Status (%s)\n\n", st.OS)

		fmt.Fprintln(out, "CPU:")
		fmt.Fprintf(out, "  %s (%d cores)\n", st.CPU.Model, st.CPU.Cores)
		fmt.Fprintf(out, "  Load Avg (1m, 5m, 15m): %.2f %.2f %.2f\n\n", st.CPU.Load1, st.CPU.Load5, st.CPU.Load15)

		fmt.Fprintln(out, "Memory:")
		fmt.Fprintf(out, "  Used: %s\n", formatGiB(st.Memory.UsedBytes))
		fmt.Fprintf(out, "  Free: %s\n", formatGiB(st.Memory.FreeBytes))
		fmt.Fprintf(out, "  Total: %s\n\n", formatGiB(st.Memory.TotalBytes))

		fmt.Fprintln(out, "Disk (/):")
		fmt.Fprintf(out, "  Used: %s\n", formatGiB(st.Disk.UsedBytes))
		fmt.Fprintf(out, "  Available: %s\n", formatGiB(st.Disk.AvailableBytes))
		fmt.Fprintf(out, "  Total: %s\n\n", formatGiB(st.Disk.TotalBytes))

		fmt.Fprintln(out, "Ports (LISTEN):")
		if len(st.Ports) == 0 {
			fmt.Fprintln(out, "  None")
		} else {
			for _, port := range st.Ports {
				fmt.Fprintf(out, "  %d  -> %s (pid %d, user %s)\n", port.Port, port.Command, port.PID, port.User)
			}
		}
		fmt.Fprintln(out)

		fmt.Fprintln(out, "Processes:")
		for _, proc := range st.Processes {
			fmt.Fprintf(out, "  %s: %d\n", proc.Name, proc.Count)
		}
		fmt.Fprintln(out)

		notes := buildNotes(st)

		fmt.Fprintln(out, "Notes:")
		if len(notes) == 0 {
			fmt.Fprintln(out, "  None")
		} else {
			for _, note := range notes {
				fmt.Fprintf(out, "  %s\n", note)
			}
		}

		return nil
	},
}

func formatGiB(bytes uint64) string {
	const gib = 1024 * 1024 * 1024
	return fmt.Sprintf("%.1fGiB", float64(bytes)/float64(gib))
}

func buildNotes(st status.Status) []string {
	notes := make([]string, 0, len(st.Notes)+2)

	if st.Disk.TotalBytes > 0 {
		freeRatio := float64(st.Disk.AvailableBytes) / float64(st.Disk.TotalBytes)
		if freeRatio < 0.10 {
			notes = append(notes, fmt.Sprintf("Low disk space: %s available", formatGiB(st.Disk.AvailableBytes)))
		}
	}

	if st.Memory.TotalBytes > 0 {
		freeRatio := float64(st.Memory.FreeBytes) / float64(st.Memory.TotalBytes)
		if freeRatio < 0.10 {
			notes = append(notes, fmt.Sprintf("Low memory: %s free", formatGiB(st.Memory.FreeBytes)))
		}
	}

	notes = append(notes, st.Notes...)
	return notes
}
