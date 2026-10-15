//go:build darwin

package status

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseLoadAvg(t *testing.T) {
	l1, l5, l15, err := parseLoadAvg("{ 1.42 1.10 0.98 }")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l1 != 1.42 || l5 != 1.10 || l15 != 0.98 {
		t.Fatalf("unexpected values: %v %v %v", l1, l5, l15)
	}

	l1, l5, l15, err = parseLoadAvg("{ 1,42 1,10 0,98 }")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l1 != 1.42 || l5 != 1.10 || l15 != 0.98 {
		t.Fatalf("unexpected values: %v %v %v", l1, l5, l15)
	}
}

func TestParseLoadAvgInvalid(t *testing.T) {
	_, _, _, err := parseLoadAvg("bad")
	if err == nil {
		t.Fatal("expected error")
	}
	expected := "status: invalid load average: \"bad\""
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestParsePageSize(t *testing.T) {
	size, ok := parsePageSize("Mach Virtual Memory Statistics: (page size of 16384 bytes)")
	if !ok {
		t.Fatal("expected ok")
	}
	if size != 16384 {
		t.Fatalf("unexpected size: %d", size)
	}
}

func TestParsePageCount(t *testing.T) {
	count, ok := parsePageCount("Pages free: 12.345.")
	if !ok {
		t.Fatal("expected ok")
	}
	if count != 12345 {
		t.Fatalf("unexpected count: %d", count)
	}
}

func TestParsePageCountInvalid(t *testing.T) {
	_, ok := parsePageCount("Pages free:")
	if ok {
		t.Fatal("expected not ok")
	}
}

func TestGetPortsTreatsNoMatchesAsEmpty(t *testing.T) {
	dir := t.TempDir()
	lsofPath := filepath.Join(dir, "lsof")
	if err := os.WriteFile(lsofPath, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("failed to write fake lsof: %v", err)
	}
	t.Setenv("PATH", dir)

	ports, err := getPorts(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ports) != 0 {
		t.Fatalf("unexpected ports: %#v", ports)
	}
}

func TestRunCommandReportsTimeout(t *testing.T) {
	dir := t.TempDir()
	slowPath := filepath.Join(dir, "slow-command")
	if err := os.WriteFile(slowPath, []byte("#!/bin/sh\n/bin/sleep 1\n"), 0o755); err != nil {
		t.Fatalf("failed to write slow command: %v", err)
	}
	t.Setenv("PATH", dir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := runCommand(ctx, "slow-command")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if err.Error() != "status: command timed out: slow-command" {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}
