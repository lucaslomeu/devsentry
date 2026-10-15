package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"devsentry/internal/clierr"
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

func TestRootCommandShowsQuickActions(t *testing.T) {
	out, err := executeRoot(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"Devsentry", "Quick actions", "devsentry docker", "devsentry clean"} {
		if !strings.Contains(out, expected) {
			t.Fatalf("output does not contain %q: %q", expected, out)
		}
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

func TestDockerHelpContainsSimpleModes(t *testing.T) {
	out, err := executeRoot(t, "docker", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, expected := range []string{"--list", "--top", "--json", "--name", "--image", "--status", "--volume"} {
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

	for _, expected := range []string{"--apply", "--images", "--containers", "--cache", "--all", "--timeout", "--older-than", "--exclude-label"} {
		if !strings.Contains(out, expected) {
			t.Fatalf("clean help does not contain %q: %q", expected, out)
		}
	}
}

func TestInvalidDockerModeIsUsageError(t *testing.T) {
	previousList, previousTop := dockerList, dockerTop
	dockerList, dockerTop = true, true
	t.Cleanup(func() {
		dockerList, dockerTop = previousList, previousTop
	})

	err := runDockerOverview(dockerCmd, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, clierr.ErrUsage) {
		t.Fatalf("unexpected error category: %v", err)
	}
}

func TestExitCodeCategories(t *testing.T) {
	if got := ExitCode(clierr.Wrap(clierr.ErrUsage, errors.New("bad flag"))); got != 2 {
		t.Fatalf("unexpected usage exit code: %d", got)
	}
	if got := ExitCode(clierr.Wrap(clierr.ErrDockerUnavailable, errors.New("offline"))); got != 3 {
		t.Fatalf("unexpected Docker exit code: %d", got)
	}
	if got := ExitCode(clierr.Wrap(clierr.ErrTimeout, errors.New("slow"))); got != 4 {
		t.Fatalf("unexpected timeout exit code: %d", got)
	}
	if got := ExitCode(clierr.Wrap(clierr.ErrMutation, errors.New("failed"))); got != 5 {
		t.Fatalf("unexpected mutation exit code: %d", got)
	}
}

func TestQuietRootSuppressesHumanOutput(t *testing.T) {
	previous := quiet
	quiet = true
	t.Cleanup(func() { quiet = previous })

	out, err := executeRoot(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Fatalf("quiet root produced output: %q", out)
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
	if got := cleanScopeArgs(false, false, false, false); got != "--images --containers --cache" {
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

func TestCommandsRejectUnexpectedArguments(t *testing.T) {
	err := statusCmd.Args(statusCmd, []string{"unexpected"})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), `unknown command "unexpected"`) {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestRootAliasesAreAvailableButHidden(t *testing.T) {
	for _, name := range []string{"ls", "summary", "top"} {
		command, _, err := rootCmd.Find([]string{name})
		if err != nil {
			t.Fatalf("failed to find alias %q: %v", name, err)
		}
		if !command.Hidden {
			t.Fatalf("alias %q should be hidden from root help", name)
		}
	}
}
