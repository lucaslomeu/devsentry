package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	sizeutil "devsentry/internal/size"
)

const maxTopItems = 10

type TopItem struct {
	Name      string `json:"name"`
	SizeBytes uint64 `json:"size_bytes"`
	SizeLabel string `json:"size"`
}

type TopOptions struct {
	IncludeAnonymousVolumes bool
	Limit                   int
	ImageFilters            []string
	VolumeFilters           []string
}

type TopResult struct {
	Images  []TopItem `json:"images"`
	Volumes []TopItem `json:"volumes"`
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
	limit := opts.Limit
	if limit <= 0 {
		limit = maxTopItems
	}

	images, err := topImages(ctx, limit, opts.ImageFilters)
	if err != nil {
		return TopResult{}, err
	}

	volumes, err := topVolumes(ctx, opts, limit)
	if err != nil {
		return TopResult{}, err
	}

	return TopResult{
		Images:  images,
		Volumes: volumes,
	}, nil
}

func topImages(ctx context.Context, limit int, filters []string) ([]TopItem, error) {
	args := []string{"image", "ls"}
	args = appendFilterArgs(args, "reference", filters)
	args = append(args, "--format", "{{json .}}")
	out, err := Run(ctx, 0, args...)
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
			return nil, fmt.Errorf("docker: invalid line %d in docker image ls: %w", lineNum, err)
		}

		name := imageDisplayName(row)
		sizeBytes, err := parseSizeToBytes(row.Size)
		if err != nil {
			return nil, fmt.Errorf("docker: invalid image size on line %d: %w", lineNum, err)
		}

		items = append(items, TopItem{
			Name:      name,
			SizeBytes: sizeBytes,
			SizeLabel: strings.TrimSpace(row.Size),
		})
	}
	sortTopItems(items)
	if len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}

func topVolumes(ctx context.Context, opts TopOptions, limit int) ([]TopItem, error) {
	args := []string{"volume", "ls"}
	args = appendFilterArgs(args, "name", opts.VolumeFilters)
	args = append(args, "--format", "{{json .}}")
	out, err := Run(ctx, 0, args...)
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
			return nil, fmt.Errorf("docker: invalid line %d in docker volume ls: %w", lineNum, err)
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
			return topVolumesFromSystemDF(ctx, volumeNames, opts, limit)
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
			return nil, fmt.Errorf("docker: invalid line %d in docker volume inspect: %w", lineNum, err)
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
	if len(items) > limit {
		items = items[:limit]
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
	return sizeutil.Parse(size)
}

func splitSize(value string) (string, string) {
	return sizeutil.Split(value)
}

func sizeUnitMultiplier(unit string) float64 {
	return sizeutil.UnitMultiplier(unit)
}

func formatBytes(bytes uint64) string {
	return sizeutil.Format(bytes)
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

func topVolumesFromSystemDF(ctx context.Context, volumeNames []string, opts TopOptions, limit int) ([]TopItem, error) {
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
			return nil, fmt.Errorf("docker: invalid line %d in docker system df: %w", lineNum, err)
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
			return nil, fmt.Errorf("docker: invalid volume size on line %d: %w", lineNum, err)
		}

		items = append(items, TopItem{
			Name:      row.Name,
			SizeBytes: sizeBytes,
			SizeLabel: strings.TrimSpace(row.Size),
		})
	}
	sortTopItems(items)
	if len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}
