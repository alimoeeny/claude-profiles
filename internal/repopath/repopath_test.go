package repopath_test

import (
	"os"
	"strings"
	"testing"

	"github.com/alimoeeny/claude-profiles/internal/repopath"
)

func TestResolve_EnvVarSet(t *testing.T) {
	t.Setenv("CLAUDE_PROFILES_DIR", "/tmp/custom-path")

	got, err := repopath.Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/tmp/custom-path" {
		t.Errorf("Resolve() = %q, want %q", got, "/tmp/custom-path")
	}
}

func TestResolve_EnvVarEmpty(t *testing.T) {
	t.Setenv("CLAUDE_PROFILES_DIR", "")

	got, err := repopath.Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, "/.claude-profiles") {
		t.Errorf("Resolve() = %q, want suffix /.claude-profiles", got)
	}
}

func TestResolve_EnvVarOverridesDefault(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	override := home + "/override-profiles"
	t.Setenv("CLAUDE_PROFILES_DIR", override)

	got, err := repopath.Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != override {
		t.Errorf("Resolve() = %q, want %q", got, override)
	}
}
