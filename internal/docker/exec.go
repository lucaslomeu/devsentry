package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const DefaultTimeout = 10 * time.Second

func Run(ctx context.Context, timeout time.Duration, args ...string) (string, error) {
	return run(ctx, timeout, nil, args...)
}

func RunWithEnv(ctx context.Context, timeout time.Duration, env []string, args ...string) (string, error) {
	return run(ctx, timeout, env, args...)
}

func run(ctx context.Context, timeout time.Duration, env []string, args ...string) (string, error) {
	effectiveTimeout := timeout
	if effectiveTimeout <= 0 {
		effectiveTimeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, effectiveTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		cmdString := fmt.Sprintf("docker %s", strings.Join(args, " "))
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("docker: not found in PATH (tried: %s)", cmdString)
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("docker: command timed out after %s: %s", effectiveTimeout, cmdString)
		}

		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		if isDaemonNotRunning(msg) {
			return "", fmt.Errorf("docker: daemon is not running: %s", cmdString)
		}

		return "", fmt.Errorf("docker: failed to run %s: %s", cmdString, msg)
	}

	return out.String(), nil
}

func isDaemonNotRunning(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "cannot connect to the docker daemon") ||
		strings.Contains(lower, "is the docker daemon running") ||
		strings.Contains(lower, "error during connect") ||
		strings.Contains(lower, "connect: no such file or directory") ||
		strings.Contains(lower, "connection refused")
}
