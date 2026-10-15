package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Summary struct {
	ContainersRunning int    `json:"containers_running"`
	ContainersStopped int    `json:"containers_stopped"`
	Images            int    `json:"images"`
	ImagesSize        string `json:"images_size"`
	Volumes           int    `json:"volumes"`
	VolumesSize       string `json:"volumes_size"`
	BuildCache        int    `json:"build_cache"`
	BuildCacheSize    string `json:"build_cache_size"`
}

type dockerInfoRow struct {
	ContainersRunning int `json:"ContainersRunning"`
	ContainersStopped int `json:"ContainersStopped"`
}

type systemDFRow struct {
	Type       string `json:"Type"`
	TotalCount string `json:"TotalCount"`
	Size       string `json:"Size"`
}

// GetSummary returns a Docker summary (containers, images, volumes, build cache).
func GetSummary(ctx context.Context) (Summary, error) {
	infoOut, err := Run(ctx, 0, "info", "--format", "{{json .}}")
	if err != nil {
		return Summary{}, err
	}

	running, stopped, err := parseDockerInfo(infoOut)
	if err != nil {
		return Summary{}, err
	}

	dfOut, err := Run(ctx, 0, "system", "df", "--format", "{{json .}}")
	if err != nil {
		return Summary{}, err
	}

	images, imagesSize, volumes, volumesSize, buildCache, buildCacheSize, err := parseSystemDF(dfOut)
	if err != nil {
		return Summary{}, err
	}

	return Summary{
		ContainersRunning: running,
		ContainersStopped: stopped,
		Images:            images,
		ImagesSize:        imagesSize,
		Volumes:           volumes,
		VolumesSize:       volumesSize,
		BuildCache:        buildCache,
		BuildCacheSize:    buildCacheSize,
	}, nil
}

func parseDockerInfo(output string) (int, int, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return 0, 0, fmt.Errorf("docker: info returned empty")
	}

	var row dockerInfoRow
	if err := json.Unmarshal([]byte(trimmed), &row); err != nil {
		return 0, 0, fmt.Errorf("docker: invalid info JSON: %w", err)
	}

	return row.ContainersRunning, row.ContainersStopped, nil
}

func parseSystemDF(output string) (int, string, int, string, int, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	lineNum := 0
	images := 0
	imagesSize := ""
	volumes := 0
	volumesSize := ""
	buildCache := 0
	buildCacheSize := ""

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var row systemDFRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return 0, "", 0, "", 0, "", fmt.Errorf("docker: invalid system df line %d: %w", lineNum, err)
		}

		countStr := strings.TrimSpace(row.TotalCount)
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return 0, "", 0, "", 0, "", fmt.Errorf("docker: invalid system df total count on line %d: %w", lineNum, err)
		}

		switch row.Type {
		case "Images":
			images = count
			imagesSize = strings.TrimSpace(row.Size)
		case "Local Volumes", "Volumes":
			volumes = count
			volumesSize = strings.TrimSpace(row.Size)
		case "Build Cache":
			buildCache = count
			buildCacheSize = strings.TrimSpace(row.Size)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, "", 0, "", 0, "", fmt.Errorf("docker: error reading system df output: %w", err)
	}

	return images, imagesSize, volumes, volumesSize, buildCache, buildCacheSize, nil
}
