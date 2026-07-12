package clean

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fakeDockerScript = `#!/bin/sh
printf '%s\n' "$*" >> "$FAKE_DOCKER_LOG"
if [ "$1" = "image" ] && [ "$2" = "ls" ]; then
  printf '%s\n' '{"ID":"sha256:abc123","Size":"10MB"}'
fi
exit 0
`

func setupFakeDocker(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	dockerPath := filepath.Join(dir, "docker")
	if err := os.WriteFile(dockerPath, []byte(fakeDockerScript), 0o755); err != nil {
		t.Fatalf("failed to write fake docker: %v", err)
	}

	logPath := filepath.Join(dir, "commands.log")
	t.Setenv("PATH", dir)
	t.Setenv("FAKE_DOCKER_LOG", logPath)
	return logPath
}

func TestParseContainerSize(t *testing.T) {
	bytes, label, err := parseContainerSize("12.3MB (virtual 50.2MB)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := uint64(123 * 1024 * 1024 / 10)
	if bytes != expected {
		t.Fatalf("unexpected bytes: %d", bytes)
	}
	if label != "12.3MB" {
		t.Fatalf("unexpected label: %q", label)
	}
}

func TestParseContainerSizeEmpty(t *testing.T) {
	bytes, label, err := parseContainerSize("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes != 0 || label != "0B" {
		t.Fatalf("unexpected result: %d %s", bytes, label)
	}
}

func TestParseReclaimedSpace(t *testing.T) {
	label, bytes := parseReclaimedSpace("Total reclaimed space: 2.1GB\n")
	if label != "2.1GB" {
		t.Fatalf("unexpected label: %q", label)
	}

	expected := uint64(21 * 1024 * 1024 * 1024 / 10)
	if bytes != expected {
		t.Fatalf("unexpected bytes: %d", bytes)
	}
}

func TestParseSizeToBytesInvalid(t *testing.T) {
	_, err := parseSizeToBytes("GB")
	if err == nil {
		t.Fatal("expected error")
	}
	expected := "invalid size: \"GB\""
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestBuildPlanOnlyCollectsSelectedScope(t *testing.T) {
	logPath := setupFakeDocker(t)

	plan, err := BuildPlan(context.Background(), Options{Images: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.DanglingImages) != 1 {
		t.Fatalf("unexpected images: %#v", plan.DanglingImages)
	}
	if len(plan.StoppedContainers) != 0 || plan.BuildCache.Available {
		t.Fatalf("unselected scopes should be empty: %#v", plan)
	}

	commands := readFakeDockerLog(t, logPath)
	if len(commands) != 1 || !strings.HasPrefix(commands[0], "image ls ") {
		t.Fatalf("unexpected commands: %#v", commands)
	}
}

func TestApplyExecutesOnlySelectedScopes(t *testing.T) {
	tests := []struct {
		name     string
		opts     Options
		expected []string
	}{
		{
			name:     "images only",
			opts:     Options{Images: true},
			expected: []string{"image prune -f"},
		},
		{
			name:     "containers and cache",
			opts:     Options{Containers: true, BuildCache: true},
			expected: []string{"container prune -f", "builder prune -f"},
		},
		{
			name:     "default remains all scopes",
			opts:     Options{},
			expected: []string{"image prune -f", "container prune -f", "builder prune -f"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logPath := setupFakeDocker(t)

			if err := Apply(context.Background(), tt.opts); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			commands := readFakeDockerLog(t, logPath)
			if strings.Join(commands, "\n") != strings.Join(tt.expected, "\n") {
				t.Fatalf("unexpected commands: %#v", commands)
			}
		})
	}
}

func readFakeDockerLog(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fake docker log: %v", err)
	}
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}
