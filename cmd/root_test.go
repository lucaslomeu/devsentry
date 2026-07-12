package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func executeRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)

	_, err := rootCmd.ExecuteC()
	return out.String(), err
}

func TestRootCommandShowsWelcome(t *testing.T) {
	out, err := executeRoot(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Welcome to devsentry! Use 'devsentry --help' to see available commands.\n"
	if out != expected {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestVersionCommandOutput(t *testing.T) {
	previous := Version
	Version = "0.1.1-test"
	t.Cleanup(func() { Version = previous })

	out, err := executeRoot(t, "version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "devsentry version: 0.1.1-test\n"
	if out != expected {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestDockerHelpContainsSubcommands(t *testing.T) {
	out, err := executeRoot(t, "docker", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"ls", "summary", "top"} {
		if !strings.Contains(out, expected) {
			t.Fatalf("docker help does not contain %q: %q", expected, out)
		}
	}
}

func TestCleanHelpShowsFlags(t *testing.T) {
	out, err := executeRoot(t, "clean", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"--apply", "--images", "--containers", "--cache", "--all", "--timeout"} {
		if !strings.Contains(out, expected) {
			t.Fatalf("clean help does not contain %q: %q", expected, out)
		}
	}
}

func TestValidateCleanFlagsRequiresScopeForApply(t *testing.T) {
	err := validateCleanFlags(true, false, false, false, false)
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "clean: --apply requires --images, --containers, --cache, or --all"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestValidateCleanFlagsRejectsAllWithIndividualScope(t *testing.T) {
	err := validateCleanFlags(false, true, true, false, false)
	if err == nil {
		t.Fatal("expected error")
	}

	expected := "clean: --all cannot be combined with --images, --containers, or --cache"
	if err.Error() != expected {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestCleanScopeArgs(t *testing.T) {
	if got := cleanScopeArgs(false, false, false, false); got != "--all" {
		t.Fatalf("unexpected default scope: %q", got)
	}
	if got := cleanScopeArgs(false, true, true, false); got != "--images --containers" {
		t.Fatalf("unexpected selected scopes: %q", got)
	}
}

func TestStatusHelp(t *testing.T) {
	out, err := executeRoot(t, "status", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "Local environment status") {
		t.Fatalf("unexpected status help output: %q", out)
	}
}

func TestRootHelpDoesNotShowCompletion(t *testing.T) {
	out, err := executeRoot(t, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, "completion") {
		t.Fatalf("root help should not contain completion command: %q", out)
	}
}
