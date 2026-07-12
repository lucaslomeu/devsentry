package clean

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"devsentry/internal/docker"
)

const defaultCleanTimeout = 60 * time.Second

type Options struct {
	Timeout    time.Duration
	Images     bool
	Containers bool
	BuildCache bool
}

type Plan struct {
	DanglingImages    []ImageItem
	StoppedContainers []ContainerItem
	BuildCache        BuildCacheInfo
	PlannedActions    []string
	EstimatedBytes    uint64
	EstimateApprox    bool
}

type ImageItem struct {
	ID        string
	SizeBytes uint64
	SizeLabel string
}

type ContainerItem struct {
	ID        string
	Name      string
	SizeBytes uint64
	SizeLabel string
}

type BuildCacheInfo struct {
	SizeBytes uint64
	SizeLabel string
	Available bool
	Approx    bool
}

type imageRow struct {
	ID   string `json:"ID"`
	Size string `json:"Size"`
}

type containerRow struct {
	ID    string `json:"ID"`
	Names string `json:"Names"`
	Size  string `json:"Size"`
}

// BuildPlan collects the items that would be cleaned and estimates reclaimable space.
func BuildPlan(ctx context.Context, opts Options) (Plan, error) {
	timeout := normalizeTimeout(opts.Timeout)
	imagesSelected, containersSelected, cacheSelected := selectedScopes(opts)

	var images []ImageItem
	var imagesTotal uint64
	if imagesSelected {
		var err error
		images, imagesTotal, err = listDanglingImages(ctx, timeout)
		if err != nil {
			return Plan{}, err
		}
	}

	var containers []ContainerItem
	var containersTotal uint64
	if containersSelected {
		var err error
		containers, containersTotal, err = listStoppedContainers(ctx, timeout)
		if err != nil {
			return Plan{}, err
		}
	}

	var buildCache BuildCacheInfo
	if cacheSelected {
		var err error
		buildCache, err = getBuildCacheEstimate(ctx, timeout)
		if err != nil {
			return Plan{}, err
		}
	}

	plannedActions := make([]string, 0, 3)
	if imagesSelected {
		plannedActions = append(plannedActions, "docker image prune")
	}
	if containersSelected {
		plannedActions = append(plannedActions, "docker container prune")
	}
	if cacheSelected {
		plannedActions = append(plannedActions, "docker builder prune")
	}

	estimated := imagesTotal + containersTotal
	estimateApprox := false
	if cacheSelected {
		if buildCache.Available {
			estimated += buildCache.SizeBytes
			if buildCache.Approx {
				estimateApprox = true
			}
		} else {
			estimateApprox = true
		}
	}

	return Plan{
		DanglingImages:    images,
		StoppedContainers: containers,
		BuildCache:        buildCache,
		PlannedActions:    plannedActions,
		EstimatedBytes:    estimated,
		EstimateApprox:    estimateApprox,
	}, nil
}

// Apply executes the cleanup commands.
func Apply(ctx context.Context, opts Options) error {
	timeout := normalizeTimeout(opts.Timeout)
	imagesSelected, containersSelected, cacheSelected := selectedScopes(opts)

	if imagesSelected {
		if _, err := docker.Run(ctx, timeout, "image", "prune", "-f"); err != nil {
			return err
		}
	}
	if containersSelected {
		if _, err := docker.Run(ctx, timeout, "container", "prune", "-f"); err != nil {
			return err
		}
	}
	if cacheSelected {
		if _, err := docker.Run(ctx, timeout, "builder", "prune", "-f"); err != nil {
			return err
		}
	}

	return nil
}

func selectedScopes(opts Options) (images, containers, cache bool) {
	if !opts.Images && !opts.Containers && !opts.BuildCache {
		return true, true, true
	}
	return opts.Images, opts.Containers, opts.BuildCache
}

func listDanglingImages(ctx context.Context, timeout time.Duration) ([]ImageItem, uint64, error) {
	out, err := docker.Run(ctx, timeout, "image", "ls", "-f", "dangling=true", "--format", "{{json .}}")
	if err != nil {
		return nil, 0, err
	}

	items := make([]ImageItem, 0)
	var total uint64
	lineNum := 0
	for _, raw := range strings.Split(out, "\n") {
		lineNum++
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		var row imageRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, 0, fmt.Errorf("clean: invalid line %d in docker image ls: %w", lineNum, err)
		}

		sizeBytes, err := parseSizeToBytes(row.Size)
		if err != nil {
			return nil, 0, fmt.Errorf("clean: invalid image size on line %d: %w", lineNum, err)
		}

		items = append(items, ImageItem{
			ID:        shortenID(row.ID),
			SizeBytes: sizeBytes,
			SizeLabel: strings.TrimSpace(row.Size),
		})
		total += sizeBytes
	}

	return items, total, nil
}

func listStoppedContainers(ctx context.Context, timeout time.Duration) ([]ContainerItem, uint64, error) {
	out, err := docker.Run(ctx, timeout, "ps", "-a", "--filter", "status=exited", "--filter", "status=dead", "--size", "--format", "{{json .}}")
	if err != nil {
		return nil, 0, err
	}

	items := make([]ContainerItem, 0)
	var total uint64
	lineNum := 0
	for _, raw := range strings.Split(out, "\n") {
		lineNum++
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		var row containerRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, 0, fmt.Errorf("clean: invalid line %d in docker ps: %w", lineNum, err)
		}

		sizeBytes, sizeLabel, err := parseContainerSize(row.Size)
		if err != nil {
			return nil, 0, fmt.Errorf("clean: invalid container size on line %d: %w", lineNum, err)
		}

		items = append(items, ContainerItem{
			ID:        shortenID(row.ID),
			Name:      strings.TrimSpace(row.Names),
			SizeBytes: sizeBytes,
			SizeLabel: sizeLabel,
		})
		total += sizeBytes
	}

	return items, total, nil
}

func getBuildCacheEstimate(ctx context.Context, timeout time.Duration) (BuildCacheInfo, error) {
	out, err := docker.Run(ctx, timeout, "builder", "prune", "--dry-run", "--force")
	if err == nil {
		sizeLabel, sizeBytes := parseReclaimedSpace(out)
		if sizeLabel != "" {
			return BuildCacheInfo{SizeBytes: sizeBytes, SizeLabel: sizeLabel, Available: true}, nil
		}
		return BuildCacheInfo{Available: false}, nil
	}

	if isUnsupportedDryRun(err) {
		return getBuildCacheFromSystemDF(ctx, timeout)
	}

	return BuildCacheInfo{}, err
}

func getBuildCacheFromSystemDF(ctx context.Context, timeout time.Duration) (BuildCacheInfo, error) {
	out, err := docker.Run(ctx, timeout, "system", "df", "--format", "{{json .}}")
	if err != nil {
		return BuildCacheInfo{}, err
	}

	lineNum := 0
	for _, raw := range strings.Split(out, "\n") {
		lineNum++
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		var row struct {
			Type string `json:"Type"`
			Size string `json:"Size"`
		}
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return BuildCacheInfo{}, fmt.Errorf("clean: invalid line %d in docker system df: %w", lineNum, err)
		}

		if row.Type != "Build Cache" {
			continue
		}

		sizeBytes, err := parseSizeToBytes(row.Size)
		if err != nil {
			return BuildCacheInfo{}, fmt.Errorf("clean: invalid build cache size on line %d: %w", lineNum, err)
		}

		return BuildCacheInfo{SizeBytes: sizeBytes, SizeLabel: strings.TrimSpace(row.Size), Available: true, Approx: true}, nil
	}

	return BuildCacheInfo{Available: false, Approx: true}, nil
}

func parseContainerSize(size string) (uint64, string, error) {
	trimmed := strings.TrimSpace(size)
	if trimmed == "" {
		return 0, "0B", nil
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return 0, "0B", nil
	}

	sizeLabel := fields[0]
	bytes, err := parseSizeToBytes(sizeLabel)
	if err != nil {
		return 0, "", err
	}

	return bytes, sizeLabel, nil
}

func parseReclaimedSpace(output string) (string, uint64) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "Total reclaimed space") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		value := strings.TrimSpace(parts[len(parts)-1])
		bytes, err := parseSizeToBytes(value)
		if err != nil {
			return value, 0
		}
		return value, bytes
	}

	return "", 0
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

func shortenID(id string) string {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return "<unknown>"
	}
	if len(trimmed) <= 12 {
		return trimmed
	}
	return trimmed[:12] + "..."
}

func isUnsupportedDryRun(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unknown flag: --dry-run") || strings.Contains(msg, "unknown flag --dry-run")
}

func normalizeTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultCleanTimeout
	}
	return timeout
}
