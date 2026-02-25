//go:build darwin

package status

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func GetStatus(ctx context.Context, _ Options) (Status, error) {
	cpu, err := getCPUInfo(ctx)
	if err != nil {
		return Status{}, err
	}

	mem, err := getMemoryInfo(ctx)
	if err != nil {
		return Status{}, err
	}

	disk, err := getDiskInfo(ctx)
	if err != nil {
		return Status{}, err
	}

	ports, err := getPorts(ctx)
	if err != nil {
		return Status{}, err
	}

	processes, err := getProcessCounts(ctx)
	if err != nil {
		return Status{}, err
	}

	notes := []string{dockerNote(ctx)}

	return Status{
		OS:        "macOS",
		CPU:       cpu,
		Memory:    mem,
		Disk:      disk,
		Ports:     ports,
		Processes: processes,
		Notes:     notes,
	}, nil
}

func getCPUInfo(ctx context.Context) (CPUInfo, error) {
	model, err := runCommand(ctx, "sysctl", "-n", "machdep.cpu.brand_string")
	if err != nil {
		return CPUInfo{}, err
	}

	coresStr, err := runCommand(ctx, "sysctl", "-n", "hw.physicalcpu")
	if err != nil {
		return CPUInfo{}, err
	}

	cores, err := strconv.Atoi(strings.TrimSpace(coresStr))
	if err != nil {
		return CPUInfo{}, fmt.Errorf("status: failed to parse CPU core count: %w", err)
	}

	loadAvg, err := runCommand(ctx, "sysctl", "-n", "vm.loadavg")
	if err != nil {
		return CPUInfo{}, err
	}

	load1, load5, load15, err := parseLoadAvg(loadAvg)
	if err != nil {
		return CPUInfo{}, err
	}

	return CPUInfo{
		Model:  strings.TrimSpace(model),
		Cores:  cores,
		Load1:  load1,
		Load5:  load5,
		Load15: load15,
	}, nil
}

func getMemoryInfo(ctx context.Context) (MemoryInfo, error) {
	totalStr, err := runCommand(ctx, "sysctl", "-n", "hw.memsize")
	if err != nil {
		return MemoryInfo{}, err
	}

	total, err := strconv.ParseUint(strings.TrimSpace(totalStr), 10, 64)
	if err != nil {
		return MemoryInfo{}, fmt.Errorf("status: failed to parse total memory: %w", err)
	}

	vmStat, err := runCommand(ctx, "vm_stat")
	if err != nil {
		return MemoryInfo{}, err
	}

	pageSize := uint64(4096)
	freePages := uint64(0)
	inactivePages := uint64(0)
	speculativePages := uint64(0)

	scanner := bufio.NewScanner(strings.NewReader(vmStat))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if lineNum == 1 {
			if size, ok := parsePageSize(line); ok {
				pageSize = size
			}
			continue
		}

		switch {
		case strings.HasPrefix(line, "Pages free:"):
			if count, ok := parsePageCount(line); ok {
				freePages = count
			}
		case strings.HasPrefix(line, "Pages inactive:"):
			if count, ok := parsePageCount(line); ok {
				inactivePages = count
			}
		case strings.HasPrefix(line, "Pages speculative:"):
			if count, ok := parsePageCount(line); ok {
				speculativePages = count
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return MemoryInfo{}, fmt.Errorf("status: failed to read vm_stat output: %w", err)
	}

	freeBytes := (freePages + inactivePages + speculativePages) * pageSize
	usedBytes := uint64(0)
	if total > freeBytes {
		usedBytes = total - freeBytes
	}

	return MemoryInfo{
		UsedBytes:  usedBytes,
		FreeBytes:  freeBytes,
		TotalBytes: total,
	}, nil
}

func getDiskInfo(ctx context.Context) (DiskInfo, error) {
	out, err := runCommand(ctx, "df", "-k", "/")
	if err != nil {
		return DiskInfo{}, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return DiskInfo{}, fmt.Errorf("status: unexpected df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return DiskInfo{}, fmt.Errorf("status: unexpected df output")
	}

	totalKiB, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return DiskInfo{}, fmt.Errorf("status: failed to parse df total: %w", err)
	}
	usedKiB, err := strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return DiskInfo{}, fmt.Errorf("status: failed to parse df used: %w", err)
	}
	availKiB, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return DiskInfo{}, fmt.Errorf("status: failed to parse df available: %w", err)
	}

	return DiskInfo{
		UsedBytes:      usedKiB * 1024,
		AvailableBytes: availKiB * 1024,
		TotalBytes:     totalKiB * 1024,
	}, nil
}

func getPorts(ctx context.Context) ([]PortInfo, error) {
	out, err := runCommand(ctx, "lsof", "-nP", "-iTCP", "-sTCP:LISTEN")
	if err != nil {
		return nil, err
	}

	portMap := parseLsofPorts(out)
	list := make([]int, 0, len(portMap))
	for port := range portMap {
		list = append(list, port)
	}
	sort.Ints(list)

	result := make([]PortInfo, 0, len(list))
	for _, port := range list {
		info := portMap[port]
		info.Port = port
		info.InUse = true
		result = append(result, info)
	}
	return result, nil
}

func getProcessCounts(ctx context.Context) ([]ProcessCount, error) {
	out, err := runCommand(ctx, "ps", "-ax", "-o", "comm=")
	if err != nil {
		return nil, err
	}

	rules := []processRule{
		{
			Label: "Java processes",
			Match: func(name string) bool { return name == "java" },
		},
		{
			Label: "Node processes",
			Match: func(name string) bool { return name == "node" },
		},
		{
			Label: "Python processes",
			Match: func(name string) bool { return name == "python" || name == "python3" },
		},
		{
			Label: "Go processes",
			Match: func(name string) bool { return name == "go" },
		},
		{
			Label: "Postgres processes",
			Match: func(name string) bool { return name == "postgres" || name == "postmaster" },
		},
		{
			Label: "Redis processes",
			Match: func(name string) bool { return name == "redis-server" },
		},
	}

	counts := make([]int, len(rules))

	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		name := strings.ToLower(filepath.Base(line))
		for i, rule := range rules {
			if rule.Match(name) {
				counts[i]++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("status: failed to read ps output: %w", err)
	}

	result := make([]ProcessCount, 0, len(rules))
	for i, rule := range rules {
		result = append(result, ProcessCount{
			Name:  rule.Label,
			Count: counts[i],
		})
	}

	return result, nil
}

type processRule struct {
	Label string
	Match func(string) bool
}

func dockerNote(ctx context.Context) string {
	if _, err := exec.LookPath("docker"); err != nil {
		return "Docker is not available"
	}

	cmd := exec.CommandContext(ctx, "docker", "info")
	if err := cmd.Run(); err != nil {
		return "Docker is not running"
	}

	return "Docker is running"
}

var portRegex = regexp.MustCompile(`:(\d+)(?:->|\s|\)|$)`) // last port is usually the listener

func parseLsofPorts(output string) map[int]PortInfo {
	ports := make(map[int]PortInfo)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "COMMAND") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		matches := portRegex.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}

		portStr := matches[len(matches)-1][1]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		if _, exists := ports[port]; exists {
			continue
		}

		pid, _ := strconv.Atoi(fields[1])
		ports[port] = PortInfo{
			Command: fields[0],
			PID:     pid,
			User:    fields[2],
			InUse:   true,
		}
	}

	return ports
}

func parsePageSize(line string) (uint64, bool) {
	re := regexp.MustCompile(`page size of (\d+) bytes`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return 0, false
	}

	size, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, false
	}

	return size, true
}

func parsePageCount(line string) (uint64, bool) {
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return 0, false
	}

	countStr := strings.TrimSpace(parts[1])
	countStr = strings.TrimSuffix(countStr, ".")
	countStr = strings.ReplaceAll(countStr, ".", "")
	if countStr == "" {
		return 0, false
	}

	count, err := strconv.ParseUint(countStr, 10, 64)
	if err != nil {
		return 0, false
	}

	return count, true
}

func parseLoadAvg(line string) (float64, float64, float64, error) {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "{")
	trimmed = strings.TrimSuffix(trimmed, "}")
	parts := strings.Fields(trimmed)
	if len(parts) < 3 {
		return 0, 0, 0, fmt.Errorf("status: invalid load average: %q", line)
	}

	load1, err := strconv.ParseFloat(strings.ReplaceAll(parts[0], ",", "."), 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("status: invalid load average: %q", line)
	}
	load5, err := strconv.ParseFloat(strings.ReplaceAll(parts[1], ",", "."), 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("status: invalid load average: %q", line)
	}
	load15, err := strconv.ParseFloat(strings.ReplaceAll(parts[2], ",", "."), 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("status: invalid load average: %q", line)
	}

	return load1, load5, load15, nil
}

func runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("status: command not found: %s", name)
		}

		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("status: failed to run %s: %s", name, msg)
	}

	return out.String(), nil
}
