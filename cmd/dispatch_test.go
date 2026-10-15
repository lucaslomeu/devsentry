package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const fakeCommandDocker = `#!/bin/sh
case "$1 $2" in
  "ps --format")
    printf '%s\n' '{"Names":"api","Image":"nginx","Status":"Up 2 minutes","Ports":"80/tcp"}'
    ;;
  "image ls")
    printf '%s\n' '{"Repository":"nginx","Tag":"latest","ID":"sha256:abc","Size":"10MB"}'
    ;;
  "volume ls")
    ;;
  "info --format")
    printf '%s\n' '{"ContainersRunning":1,"ContainersStopped":0}'
    ;;
  "system df")
    printf '%s\n' '{"Type":"Images","TotalCount":"1","Size":"10MB"}'
    printf '%s\n' '{"Type":"Local Volumes","TotalCount":"0","Size":"0B"}'
    printf '%s\n' '{"Type":"Build Cache","TotalCount":"0","Size":"0B"}'
    ;;
esac
exit 0
`

func setupCommandDocker(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "docker")
	if err := os.WriteFile(path, []byte(fakeCommandDocker), 0o755); err != nil {
		t.Fatalf("failed to write fake docker: %v", err)
	}
	t.Setenv("PATH", dir)
}

func resetCommandState() {
	dockerList = false
	dockerTop = false
	dockerOverviewJSON = false
	dockerOverviewAll = false
	dockerOverviewLimit = 10
	dockerFilterNames = nil
	dockerFilterImages = nil
	dockerFilterStatuses = nil
	dockerFilterVolumes = nil
	dockerLsJSON = false
	dockerSummaryJSON = false
	dockerTopJSON = false
	dockerTopAll = false
	dockerTopLimit = 10
	cleanApply = false
	cleanImages = false
	cleanContainers = false
	cleanCache = false
	cleanAll = false
	cleanJSON = false
	cleanOlderThan = 0
	cleanExcludeLabels = nil
	statusJSON = false
	quiet = false
}

func TestDockerModesDispatchToJSONOutput(t *testing.T) {
	setupCommandDocker(t)

	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "summary", args: []string{"docker", "--json"}, want: "docker"},
		{name: "list", args: []string{"docker", "--list", "--json", "--name", "api"}, want: "docker.list"},
		{name: "top", args: []string{"docker", "--top", "--json", "--image", "nginx:*"}, want: "docker.top"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCommandState()
			out, err := executeRoot(t, tt.args...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var envelope JSONEnvelope
			if err := json.Unmarshal([]byte(out), &envelope); err != nil {
				t.Fatalf("invalid JSON output: %v (%q)", err, out)
			}
			if envelope.SchemaVersion != 1 || envelope.Command != tt.want {
				t.Fatalf("unexpected envelope: %#v", envelope)
			}
		})
	}
}
