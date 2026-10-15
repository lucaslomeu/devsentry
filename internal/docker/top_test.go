package docker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseSizeToBytes(t *testing.T) {
	bytes, err := parseSizeToBytes("1KB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes != 1024 {
		t.Fatalf("unexpected bytes: %d", bytes)
	}

	bytes, err = parseSizeToBytes("1.5GB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := uint64(1.5 * 1024 * 1024 * 1024)
	if bytes != expected {
		t.Fatalf("unexpected bytes: %d", bytes)
	}

	bytes, err = parseSizeToBytes("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes != 0 {
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

	_, err = parseSizeToBytes("12XB")
	if err == nil {
		t.Fatal("expected error")
	}
	expected = "unknown size unit: \"XB\""
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestFormatBytes(t *testing.T) {
	if got := formatBytes(0); got != "0B" {
		t.Fatalf("unexpected format: %q", got)
	}
	if got := formatBytes(1536); got != "1.5KiB" {
		t.Fatalf("unexpected format: %q", got)
	}
}

func TestImageDisplayName(t *testing.T) {
	name := imageDisplayName(dockerImageRow{Repository: "<none>", Tag: "<none>", ID: "abc"})
	if name != "abc" {
		t.Fatalf("unexpected name: %q", name)
	}

	name = imageDisplayName(dockerImageRow{Repository: "postgres", Tag: "<none>", ID: "abc"})
	if name != "postgres" {
		t.Fatalf("unexpected name: %q", name)
	}

	name = imageDisplayName(dockerImageRow{Repository: "myapp", Tag: "latest", ID: "abc"})
	if name != "myapp:latest" {
		t.Fatalf("unexpected name: %q", name)
	}

	name = imageDisplayName(dockerImageRow{Repository: "", Tag: "", ID: ""})
	if name != "<unknown>" {
		t.Fatalf("unexpected name: %q", name)
	}
}

func TestSplitSize(t *testing.T) {
	value, unit := splitSize("12.3GB")
	if value != "12.3" || unit != "GB" {
		t.Fatalf("unexpected split: %s %s", value, unit)
	}

	value, unit = splitSize("100")
	if value != "100" || unit != "" {
		t.Fatalf("unexpected split: %s %s", value, unit)
	}
}

func TestSizeUnitMultiplier(t *testing.T) {
	if got := sizeUnitMultiplier("KB"); got != 1024 {
		t.Fatalf("unexpected multiplier: %v", got)
	}
	if got := sizeUnitMultiplier("mib"); got != 1024*1024 {
		t.Fatalf("unexpected multiplier: %v", got)
	}
}

func TestGetTopHonorsLimit(t *testing.T) {
	dir := t.TempDir()
	dockerPath := filepath.Join(dir, "docker")
	script := `#!/bin/sh
if [ "$1 $2" = "image ls" ]; then
  printf '%s\n' '{"Repository":"one","Tag":"latest","ID":"1","Size":"3GB"}'
  printf '%s\n' '{"Repository":"two","Tag":"latest","ID":"2","Size":"2GB"}'
  printf '%s\n' '{"Repository":"three","Tag":"latest","ID":"3","Size":"1GB"}'
fi
`
	if err := os.WriteFile(dockerPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake docker: %v", err)
	}
	t.Setenv("PATH", dir)

	result, err := GetTop(context.Background(), TopOptions{Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Images) != 2 {
		t.Fatalf("unexpected images: %#v", result.Images)
	}
	if result.Images[0].Name != "one:latest" || result.Images[1].Name != "two:latest" {
		t.Fatalf("unexpected ordering: %#v", result.Images)
	}
}
