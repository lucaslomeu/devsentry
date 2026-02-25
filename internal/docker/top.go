package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const maxTopItems = 10

type TopItem struct {
	Name      string
	SizeBytes uint64
	SizeLabel string
}

type TopOptions struct {
	IncludeAnonymousVolumes bool
}

type TopResult struct {
	Images  []TopItem
	Volumes []TopItem
}

type dockerImageRow struct {
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	ID         string `json:"ID"`
	Size       string `json:"Size"`
}

type dockerVolumeRow struct {
	Name string `json:"Name"`
}

type dockerVolumeInspect struct {
	Name      string           `json:"Name"`
	UsageData dockerUsageStats `json:"UsageData"`
}

type dockerUsageStats struct {
	Size int64 `json:"Size"`
}

type dockerSystemDFRow struct {
	Type string `json:"Type"`
	Name string `json:"Name"`
	Size string `json:"Size"`
}

// GetTop returns the top images and volumes by size.
func GetTop(ctx context.Context, opts TopOptions) (TopResult, error) {
	images, err := topImages(ctx)
	if err != nil {
		return TopResult{}, err
	}

	volumes, err := topVolumes(ctx, opts)
	if err != nil {
		return TopResult{}, err
	}

	return TopResult{
		Images:  images,
		Volumes: volumes,
	}, nil
}

func topImages(ctx context.Context) ([]TopItem, error) {
	out, err := Run(ctx, 0, "image", "ls", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}

	items := make([]TopItem, 0)
	lineNum := 0
	for _, raw := range strings.Split(out, "\n") {
		lineNum++
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		var row dockerImageRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("invalid line %d in docker image ls: %w", lineNum, err)
		}

		name := imageDisplayName(row)
		sizeBytes, err := parseSizeToBytes(row.Size)
		if err != nil {
			return nil, fmt.Errorf("invalid image size on line %d: %w", lineNum, err)
		}

		items = append(items, TopItem{
			Name:      name,
			SizeBytes: sizeBytes,
			SizeLabel: strings.TrimSpace(row.Size),
		})
	}
	sortTopItems(items)
	if len(items) > maxTopItems {
		items = items[:maxTopItems]
	}

	return items, nil
}

func topVolumes(ctx context.Context, opts TopOptions) ([]TopItem, error) {
	out, err := Run(ctx, 0, "volume", "ls", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}

	volumeNames := make([]string, 0)
	lineNum := 0
	for _, raw := range strings.Split(out, "\n") {
		lineNum++
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		var row dockerVolumeRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("invalid line %d in docker volume ls: %w", lineNum, err)
		}

		if row.Name == "" {
			continue
		}

		volumeNames = append(volumeNames, row.Name)
	}
	if !opts.IncludeAnonymousVolumes {
		volumeNames = filterNamedVolumes(volumeNames)
	}

	if len(volumeNames) == 0 {
		return []TopItem{}, nil
	}

	inspectArgs := append([]string{"volume", "inspect", "--size", "--format", "{{json .}}"}, volumeNames...)
	inspectOut, err := Run(ctx, 0, inspectArgs...)
	if err != nil {
		if isUnsupportedVolumeSizeFlag(err) {
			return topVolumesFromSystemDF(ctx, volumeNames, opts)
		}
		return nil, err
	}

	items := make([]TopItem, 0, len(volumeNames))
	lineNum = 0
	for _, raw := range strings.Split(inspectOut, "\n") {
		lineNum++
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		var row dockerVolumeInspect
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("invalid line %d in docker volume inspect: %w", lineNum, err)
		}

		if row.Name == "" {
			continue
		}

		sizeBytes := uint64(0)
		if row.UsageData.Size > 0 {
			sizeBytes = uint64(row.UsageData.Size)
		}

		items = append(items, TopItem{
			Name:      row.Name,
			SizeBytes: sizeBytes,
			SizeLabel: formatBytes(sizeBytes),
		})
	}
	sortTopItems(items)
	if len(items) > maxTopItems {
		items = items[:maxTopItems]
	}

	return items, nil
}

func imageDisplayName(row dockerImageRow) string {
	repo := strings.TrimSpace(row.Repository)
	tag := strings.TrimSpace(row.Tag)
	if repo == "" || repo == "<none>" {
		if row.ID != "" {
			return row.ID
		}
		return "<unknown>"
	}
	if tag == "" || tag == "<none>" {
		return repo
	}
	return fmt.Sprintf("%s:%s", repo, tag)
}

func sortTopItems(items []TopItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].SizeBytes == items[j].SizeBytes {
			return items[i].Name < items[j].Name
		}
		return items[i].SizeBytes > items[j].SizeBytes
	})
}

func parseSizeToBytes(size string) (uint64, error) {
	trimmed := strings.TrimSpace(size)
	if trimmed == "" {
		return 0, nil
	}

	trimmed = strings.ReplaceAll(trimmed, " ", "")

	value, unit := splitSize(trimmed)
	if value == "" {
		return 0, fmt.Errorf("invalid size: %q", size)
	}

	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size: %q", size)
	}

	multiplier := sizeUnitMultiplier(unit)
	if multiplier == 0 {
		return 0, fmt.Errorf("unknown size unit: %q", unit)
	}

	return uint64(num * multiplier), nil
}

func splitSize(value string) (string, string) {
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

func sizeUnitMultiplier(unit string) float64 {
	unit = strings.ToLower(strings.TrimSpace(unit))
	switch unit {
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

func filterNamedVolumes(names []string) []string {
	filtered := make([]string, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		if isAnonymousVolume(name) {
			continue
		}
		filtered = append(filtered, name)
	}
	return filtered
}

func isAnonymousVolume(name string) bool {
	if len(name) != 64 {
		return false
	}
	for _, r := range name {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		default:
			return false
		}
	}
	return true
}

func isUnsupportedVolumeSizeFlag(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unknown flag: --size") || strings.Contains(msg, "unknown flag --size")
}

func topVolumesFromSystemDF(ctx context.Context, volumeNames []string, opts TopOptions) ([]TopItem, error) {
	out, err := Run(ctx, 0, "system", "df", "-v", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}

	nameSet := make(map[string]struct{}, len(volumeNames))
	for _, name := range volumeNames {
		if name == "" {
			continue
		}
		if !opts.IncludeAnonymousVolumes && isAnonymousVolume(name) {
			continue
		}
		nameSet[name] = struct{}{}
	}

	items := make([]TopItem, 0)
	lineNum := 0
	for _, raw := range strings.Split(out, "\n") {
		lineNum++
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		var row dockerSystemDFRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("invalid line %d in docker system df: %w", lineNum, err)
		}

		if row.Type != "Local Volumes" && row.Type != "Volumes" {
			continue
		}
		if row.Name == "" {
			continue
		}
		if _, ok := nameSet[row.Name]; !ok {
			continue
		}

		sizeBytes, err := parseSizeToBytes(row.Size)
		if err != nil {
			return nil, fmt.Errorf("invalid volume size on line %d: %w", lineNum, err)
		}

		items = append(items, TopItem{
			Name:      row.Name,
			SizeBytes: sizeBytes,
			SizeLabel: strings.TrimSpace(row.Size),
		})
	}
	sortTopItems(items)
	if len(items) > maxTopItems {
		items = items[:maxTopItems]
	}

	return items, nil
}
