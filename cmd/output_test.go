package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestOutputStyleAlwaysUsesColor(t *testing.T) {
	var out bytes.Buffer
	if style := newOutputStyle(&out); !style.color || !strings.Contains(style.success("ok"), "\033[") {
		t.Fatal("output should always use colors")
	}
}

func TestWriteCommandJSONUsesVersionedEnvelope(t *testing.T) {
	var out bytes.Buffer
	if err := writeCommandJSON(&out, "status", map[string]string{"state": "ok"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var envelope JSONEnvelope
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if envelope.SchemaVersion != 1 || envelope.Command != "status" {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
	if strings.Contains(out.String(), "\033[") {
		t.Fatalf("JSON contains ANSI escape sequence: %q", out.String())
	}
}

func TestUsageBar(t *testing.T) {
	if got := usageBar(50, 100, 10); got != "[#####-----] 50%" {
		t.Fatalf("unexpected usage bar: %q", got)
	}
	if got := usageBar(0, 0, 10); got != "[----------] n/a" {
		t.Fatalf("unexpected empty usage bar: %q", got)
	}
	if got := usageBar(200, 100, 10); got != "[##########] 100%" {
		t.Fatalf("unexpected clamped usage bar: %q", got)
	}
}
