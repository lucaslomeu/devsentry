package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type ContainerInfo struct {
	Name   string
	Image  string
	Status string
	Ports  string
}

type dockerPSRow struct {
	Names  string `json:"Names"`
	Image  string `json:"Image"`
	Status string `json:"Status"`
	Ports  string `json:"Ports"`
}

func ListRunningContainers(ctx context.Context) ([]ContainerInfo, error) {
	output, err := Run(ctx, 0, "ps", "--format", "{{json .}}")
	if err != nil {
		return nil, err
	}

	results, err := parseDockerPSOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse docker ps output: %w", err)
	}

	return results, nil
}

func parseDockerPSOutput(output string) ([]ContainerInfo, error) {
	results := make([]ContainerInfo, 0)
	scanner := bufio.NewScanner(strings.NewReader(output))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var row dockerPSRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("invalid line %d in docker ps: %w", lineNum, err)
		}

		results = append(results, ContainerInfo{
			Name:   row.Names,
			Image:  row.Image,
			Status: row.Status,
			Ports:  row.Ports,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading docker ps output: %w", err)
	}

	return results, nil
}
