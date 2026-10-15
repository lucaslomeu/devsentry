package clean

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"devsentry/internal/docker"
	sizeutil "devsentry/internal/size"
)

const defaultCleanTimeout = 60 * time.Second

type Options struct {
	Timeout    time.Duration
	Images     bool
	Containers bool
	BuildCache bool
	All        bool
}

type Plan struct {
	DanglingImages    []ImageItem     `json:"dangling_images"`
	UnusedImages      []ImageItem     `json:"unused_images"`
	StoppedContainers []ContainerItem `json:"stopped_containers"`
	BuildCache        BuildCacheInfo  `json:"build_cache"`
	PlannedActions    []string        `json:"planned_actions"`
	EstimatedBytes    uint64          `json:"estimated_bytes"`
	EstimateApprox    bool            `json:"estimate_approx"`
}

type ImageItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SizeBytes uint64 `json:"size_bytes"`
	SizeLabel string `json:"size"`
}

type ContainerItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SizeBytes uint64 `json:"size_bytes"`
	SizeLabel string `json:"size"`
}

type BuildCacheInfo struct {
	SizeBytes uint64 `json:"size_bytes"`
	SizeLabel string `json:"size"`
	Available bool   `json:"available"`
	Approx    bool   `json:"approx"`
}

type imageRow struct {
	ID         string `json:"ID"`
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	Size       string `json:"Size"`
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
	var unusedImages []ImageItem
	var imagesTotal uint64
	if imagesSelected {
		var err error
		images, unusedImages, imagesTotal, err = listImageCleanupCandidates(ctx, timeout, opts.All, containersSelected)
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
	if containersSelected {
		plannedActions = append(plannedActions, "docker container prune")
	}
	if imagesSelected {
		action := "docker image prune"
		if opts.All {
			action += " -a"
		}
		plannedActions = append(plannedActions, action)
	}
	if cacheSelected {
		plannedActions = append(plannedActions, "docker builder prune")
	}

	estimated := imagesTotal + containersTotal
	estimateApprox := opts.All
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
		UnusedImages:      unusedImages,
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

	if containersSelected {
		if _, err := docker.Run(ctx, timeout, "container", "prune", "-f"); err != nil {
			return err
		}
	}
	if imagesSelected {
		args := []string{"image", "prune", "-f"}
		if opts.All {
			args = []string{"image", "prune", "-a", "-f"}
		}
		if _, err := docker.Run(ctx, timeout, args...); err != nil {
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

func listImageCleanupCandidates(ctx context.Context, timeout time.Duration, includeUnused, containersWillBePruned bool) ([]ImageItem, []ImageItem, uint64, error) {
	out, err := docker.Run(ctx, timeout, "image", "ls", "-a", "--no-trunc", "--format", "{{json .}}")
	if err != nil {
		return nil, nil, 0, err
	}

	rowsByID := make(map[string]imageRow)
	namesByID := make(map[string][]string)
	lineNum := 0
	for _, raw := range strings.Split(out, "\n") {
		lineNum++
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		var row imageRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, nil, 0, fmt.Errorf("clean: invalid line %d in docker image ls: %w", lineNum, err)
		}
		id := normalizeImageID(row.ID)
		if id == "" {
			continue
		}
		rowsByID[id] = row
		if name := imageName(row); name != "<dangling>" {
			namesByID[id] = appendUnique(namesByID[id], name)
		}
	}

	referenced, err := referencedImageIDs(ctx, timeout, !containersWillBePruned)
	if err != nil {
		return nil, nil, 0, err
	}

	dangling := make([]ImageItem, 0)
	unused := make([]ImageItem, 0)
	var total uint64
	for id, row := range rowsByID {
		if referenced[id] {
			continue
		}
		sizeBytes, err := parseSizeToBytes(row.Size)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("clean: invalid image size for %s: %w", shortenID(id), err)
		}
		item := ImageItem{ID: shortenID(id), Name: imageName(row), SizeBytes: sizeBytes, SizeLabel: strings.TrimSpace(row.Size)}
		if names := namesByID[id]; len(names) > 0 {
			item.Name = strings.Join(names, ", ")
		}
		if isDanglingImage(row) {
			dangling = append(dangling, item)
			total += sizeBytes
		} else if includeUnused {
			unused = append(unused, item)
			total += sizeBytes
		}
	}

	sortImageItems(dangling)
	sortImageItems(unused)
	return dangling, unused, total, nil
}

func referencedImageIDs(ctx context.Context, timeout time.Duration, includeStopped bool) (map[string]bool, error) {
	args := []string{"ps"}
	if includeStopped {
		args = append(args, "-a")
	}
	args = append(args, "-q")
	out, err := docker.Run(ctx, timeout, args...)
	if err != nil {
		return nil, err
	}
	containerIDs := strings.Fields(out)
	result := make(map[string]bool)
	if len(containerIDs) == 0 {
		return result, nil
	}
	inspectArgs := append([]string{"container", "inspect", "--format", "{{.Image}}"}, containerIDs...)
	out, err = docker.Run(ctx, timeout, inspectArgs...)
	if err != nil {
		return nil, err
	}
	for _, id := range strings.Fields(out) {
		if normalized := normalizeImageID(id); normalized != "" {
			result[normalized] = true
		}
	}
	return result, nil
}

func normalizeImageID(id string) string {
	return strings.TrimPrefix(strings.TrimSpace(id), "sha256:")
}

func isDanglingImage(row imageRow) bool {
	return (row.Repository == "" || row.Repository == "<none>") && (row.Tag == "" || row.Tag == "<none>")
}

func imageName(row imageRow) string {
	if isDanglingImage(row) {
		return "<dangling>"
	}
	if row.Tag == "" || row.Tag == "<none>" {
		return row.Repository
	}
	return row.Repository + ":" + row.Tag
}

func appendUnique(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func sortImageItems(items []ImageItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].SizeBytes == items[j].SizeBytes {
			return items[i].Name < items[j].Name
		}
		return items[i].SizeBytes > items[j].SizeBytes
	})
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
	return sizeutil.Parse(size)
}

func splitSize(value string) (string, string) {
	return sizeutil.Split(value)
}

func sizeUnitMultiplier(unit string) float64 {
	return sizeutil.UnitMultiplier(unit)
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
