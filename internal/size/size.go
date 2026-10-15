package size

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse converts Docker's human-readable size values into bytes.
func Parse(value string) (uint64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, nil
	}

	trimmed = strings.ReplaceAll(trimmed, " ", "")
	number, unit := Split(trimmed)
	if number == "" {
		return 0, fmt.Errorf("invalid size: %q", value)
	}

	num, err := strconv.ParseFloat(number, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size: %q", value)
	}

	multiplier := UnitMultiplier(unit)
	if multiplier == 0 {
		return 0, fmt.Errorf("unknown size unit: %q", unit)
	}

	return uint64(num * multiplier), nil
}

// Split separates a numeric size from its unit.
func Split(value string) (string, string) {
	for i, r := range value {
		if r < '0' || r > '9' {
			if r == '.' {
				continue
			}
			return value[:i], value[i:]
		}
	}
	return value, ""
}

// UnitMultiplier returns the binary multiplier for a size unit.
func UnitMultiplier(unit string) float64 {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "b", "bytes", "":
		return 1
	case "kb", "kib":
		return 1024
	case "mb", "mib":
		return 1024 * 1024
	case "gb", "gib":
		return 1024 * 1024 * 1024
	case "tb", "tib":
		return 1024 * 1024 * 1024 * 1024
	default:
		return 0
	}
}

// Format renders bytes using binary units.
func Format(bytes uint64) string {
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
