package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
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

// ListRunningContainers returns running Docker containers using `docker ps`.
func ListRunningContainers(ctx context.Context) ([]ContainerInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "--format", "{{json .}}")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("docker nao encontrado no PATH")
		}

		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("falha ao executar docker ps: %s", msg)
	}

	results := make([]ContainerInfo, 0)
	scanner := bufio.NewScanner(&out)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var row dockerPSRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return nil, fmt.Errorf("linha %d invalida no docker ps: %w", lineNum, err)
		}

		results = append(results, ContainerInfo{
			Name:   row.Names,
			Image:  row.Image,
			Status: row.Status,
			Ports:  row.Ports,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro lendo saida do docker ps: %w", err)
	}

	return results, nil
}
