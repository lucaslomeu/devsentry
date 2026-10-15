package cmd

import (
	"fmt"
	"strings"
	"time"

	"devsentry/internal/status"
	"github.com/spf13/cobra"
)

var statusJSON bool
var statusTimeout = 10 * time.Second

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Local environment status",
	Args:  func(cmd *cobra.Command, args []string) error { return usageArgs("status", args) },
	RunE:  runStatusCmd,
}

func runStatusCmd(cmd *cobra.Command, args []string) error {
	st, err := status.GetStatus(cmd.Context(), status.Options{Timeout: statusTimeout})
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if statusJSON {
		return writeCommandJSON(out, "status", st)
	}
	if quiet {
		return nil
	}

	notes := buildNotes(st)
	style := newOutputStyle(out)
	fmt.Fprintf(out, "%s\n", style.title(fmt.Sprintf("Devsentry - Environment Status (%s)", st.OS)))
	fmt.Fprintf(out, "Health: %s\n\n", styleStatus(style, statusHealth(notes)))

	fmt.Fprintln(out, style.section("CPU"))
	fmt.Fprintf(out, "  %s (%d cores)\n", st.CPU.Model, st.CPU.Cores)
	fmt.Fprintf(out, "  Load Avg (1m, 5m, 15m): %.2f %.2f %.2f\n\n", st.CPU.Load1, st.CPU.Load5, st.CPU.Load15)

	fmt.Fprintln(out, style.section("Memory"))
	fmt.Fprintf(out, "  Used: %s\n", formatGiB(st.Memory.UsedBytes))
	fmt.Fprintf(out, "  Free: %s\n", formatGiB(st.Memory.FreeBytes))
	fmt.Fprintf(out, "  Total: %s\n", formatGiB(st.Memory.TotalBytes))
	fmt.Fprintf(out, "  Usage: %s\n\n", styleUsage(style, st.Memory.UsedBytes, st.Memory.TotalBytes))

	fmt.Fprintln(out, style.section("Disk (/)"))
	fmt.Fprintf(out, "  Used: %s\n", formatGiB(st.Disk.UsedBytes))
	fmt.Fprintf(out, "  Available: %s\n", formatGiB(st.Disk.AvailableBytes))
	fmt.Fprintf(out, "  Total: %s\n", formatGiB(st.Disk.TotalBytes))
	fmt.Fprintf(out, "  Usage: %s\n\n", styleUsage(style, st.Disk.UsedBytes, st.Disk.TotalBytes))

	fmt.Fprintln(out, style.section("Ports (LISTEN)"))
	if len(st.Ports) == 0 {
		fmt.Fprintln(out, "  None")
	} else {
		for _, port := range st.Ports {
			fmt.Fprintf(out, "  %d  -> %s (pid %d, user %s)\n", port.Port, port.Command, port.PID, port.User)
		}
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, style.section("Processes"))
	for _, proc := range st.Processes {
		fmt.Fprintf(out, "  %s: %d\n", proc.Name, proc.Count)
	}
	fmt.Fprintln(out)

	fmt.Fprintln(out, style.section("Notes"))
	if len(notes) == 0 {
		fmt.Fprintln(out, "  None")
	} else {
		for _, note := range notes {
			fmt.Fprintf(out, "  %s\n", styleNote(style, note))
		}
	}

	return nil
}

func statusHealth(notes []string) string {
	for _, note := range notes {
		if strings.Contains(note, "Docker is not running") || strings.Contains(note, "Docker is not available") {
			return "CRITICAL"
		}
	}
	for _, note := range notes {
		if strings.HasPrefix(note, "Low ") {
			return "WARNING"
		}
	}
	return "OK"
}

func styleStatus(style outputStyle, value string) string {
	switch value {
	case "OK":
		return style.success(value)
	case "WARNING":
		return style.warning(value)
	default:
		return style.error(value)
	}
}

func styleUsage(style outputStyle, used, total uint64) string {
	if total == 0 {
		return style.muted(usageBar(used, total, 20))
	}
	ratio := float64(used) / float64(total)
	switch {
	case ratio >= 0.90:
		return style.error(usageBar(used, total, 20))
	case ratio >= 0.75:
		return style.warning(usageBar(used, total, 20))
	default:
		return style.success(usageBar(used, total, 20))
	}
}

func styleNote(style outputStyle, note string) string {
	switch {
	case strings.HasPrefix(note, "Low "):
		return style.warning(note)
	case strings.Contains(note, "not running"), strings.Contains(note, "not available"):
		return style.error(note)
	case strings.Contains(note, "is running"):
		return style.success(note)
	default:
		return note
	}
}

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output JSON")
	statusCmd.Flags().DurationVar(&statusTimeout, "timeout", 10*time.Second, "Timeout for status checks")
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
