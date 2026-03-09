package docker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const fakeDockerScript = `#!/bin/sh
if [ -n "$FAKE_DOCKER_SLEEP" ]; then
  /bin/sleep "$FAKE_DOCKER_SLEEP"
fi
if [ -n "$FAKE_DOCKER_STDOUT" ]; then
  printf "%s" "$FAKE_DOCKER_STDOUT"
fi
if [ -n "$FAKE_DOCKER_STDERR" ]; then
  printf "%s" "$FAKE_DOCKER_STDERR" 1>&2
fi
if [ -n "$FAKE_DOCKER_EXIT" ]; then
  exit "$FAKE_DOCKER_EXIT"
fi
exit 0
`

func writeFakeDocker(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "docker")
	if err := os.WriteFile(path, []byte(fakeDockerScript), 0o755); err != nil {
		t.Fatalf("failed to write fake docker: %v", err)
	}
	return dir
}

func TestRunDockerNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := Run(context.Background(), 0, "version")
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "docker: not found in PATH (tried: docker version)"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestRunTimeout(t *testing.T) {
	dir := writeFakeDocker(t)
	t.Setenv("PATH", dir)
	t.Setenv("FAKE_DOCKER_SLEEP", "0.2")

	timeout := 50 * time.Millisecond
	_, err := Run(context.Background(), timeout, "ps")
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "docker: command timed out after 50ms: docker ps"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestRunDaemonNotRunning(t *testing.T) {
	dir := writeFakeDocker(t)
	t.Setenv("PATH", dir)
	t.Setenv("FAKE_DOCKER_STDERR", "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?")
	t.Setenv("FAKE_DOCKER_EXIT", "1")

	_, err := Run(context.Background(), 0, "ps")
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "docker: daemon is not running: docker ps"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestRunGenericFailure(t *testing.T) {
	dir := writeFakeDocker(t)
	t.Setenv("PATH", dir)
	t.Setenv("FAKE_DOCKER_STDERR", "boom")
	t.Setenv("FAKE_DOCKER_EXIT", "1")

	_, err := Run(context.Background(), 0, "ps")
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "docker: failed to run docker ps: boom"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestRunSuccess(t *testing.T) {
	dir := writeFakeDocker(t)
	t.Setenv("PATH", dir)
	t.Setenv("FAKE_DOCKER_STDOUT", "ok")

	out, err := Run(context.Background(), 0, "ps")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Fatalf("unexpected stdout: %q", out)
	}
}

func TestRunWithEnv(t *testing.T) {
	dir := writeFakeDocker(t)
	t.Setenv("PATH", dir)

	out, err := RunWithEnv(context.Background(), 0, []string{"FAKE_DOCKER_STDOUT=env-ok"}, "ps")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "env-ok" {
		t.Fatalf("unexpected stdout: %q", out)
	}
}
