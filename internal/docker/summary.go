package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Summary struct {
	ContainersRunning int
	ContainersStopped int
	Images            int
	ImagesSize        string
	Volumes           int
	VolumesSize       string
	BuildCache        int
	BuildCacheSize    string
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
	infoOut, err := runDockerCommand(ctx, "info", "--format", "{{json .}}")
	if err != nil {
		return Summary{}, err
	}

	running, stopped, err := parseDockerInfo(infoOut)
	if err != nil {
		return Summary{}, err
	}

	dfOut, err := runDockerCommand(ctx, "system", "df", "--format", "{{json .}}")
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

func runDockerCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", args...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("docker not found: %w", err)
		}

		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("failed to execute docker %s: %s", strings.Join(args, " "), msg)
	}

	return out.String(), nil
}

func parseDockerInfo(output string) (int, int, error) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return 0, 0, fmt.Errorf("docker info returned empty")
	}

	var row dockerInfoRow
	if err := json.Unmarshal([]byte(trimmed), &row); err != nil {
		return 0, 0, fmt.Errorf("invalid docker info JSON: %w", err)
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
			return 0, "", 0, "", 0, "", fmt.Errorf("invalid line %d in docker system df: %w", lineNum, err)
		}

		countStr := strings.TrimSpace(row.TotalCount)
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return 0, "", 0, "", 0, "", fmt.Errorf("invalid TotalCount on line %d: %w", lineNum, err)
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
		return 0, "", 0, "", 0, "", fmt.Errorf("error reading docker system df output: %w", err)
	}

	return images, imagesSize, volumes, volumesSize, buildCache, buildCacheSize, nil
}
