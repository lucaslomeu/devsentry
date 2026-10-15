package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestOutputStyleAlwaysUsesColor(t *testing.T) {
	var out bytes.Buffer
	if style := newOutputStyle(&out); !style.color || !strings.Contains(style.success("ok"), "\033[") {
		t.Fatal("output should always use colors")
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
