package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type outputStyle struct {
	color bool
}

func newOutputStyle(out io.Writer) outputStyle {
	return outputStyle{color: true}
}

func (s outputStyle) header(value string) string {
	return s.section(value)
}

func (s outputStyle) title(value string) string {
	if !s.color {
		return value
	}
	return "\033[1;36m" + value + "\033[0m"
}

func (s outputStyle) section(value string) string {
	if !s.color {
		return value
	}
	return "\033[1;36m" + value + "\033[0m"
}

func (s outputStyle) success(value string) string {
	return s.paint(value, "\033[1;32m")
}

func (s outputStyle) info(value string) string {
	return s.paint(value, "\033[34m")
}

func (s outputStyle) warning(value string) string {
	return s.paint(value, "\033[1;33m")
}

func (s outputStyle) error(value string) string {
	return s.paint(value, "\033[1;31m")
}

func (s outputStyle) muted(value string) string {
	return s.paint(value, "\033[2m")
}

func (s outputStyle) state(value string) string {
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(lower, "up"), strings.Contains(lower, "running"), strings.Contains(lower, "active"):
		return s.success(value)
	case strings.Contains(lower, "exited"), strings.Contains(lower, "dead"), strings.Contains(lower, "down"), strings.Contains(lower, "error"):
		return s.error(value)
	default:
		return s.warning(value)
	}
}

func (s outputStyle) paint(value, code string) string {
	if !s.color {
		return value
	}
	return code + value + "\033[0m"
}

func usageBar(used, total uint64, width int) string {
	if width < 1 {
		width = 1
	}
	if total == 0 {
		return fmt.Sprintf("[%s] n/a", strings.Repeat("-", width))
	}

	ratio := float64(used) / float64(total)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	return fmt.Sprintf("[%s%s] %d%%", strings.Repeat("#", filled), strings.Repeat("-", width-filled), int(ratio*100))
}

func writeJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
