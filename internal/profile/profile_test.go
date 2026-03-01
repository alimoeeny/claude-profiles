package profile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ali/claude-profile-switcher/internal/profile"
)

func TestSnapshot(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()

	os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"version":1}`), 0644)
	os.MkdirAll(filepath.Join(home, ".claude", "mcp"), 0755)
	os.WriteFile(filepath.Join(home, ".claude", "settings.json"), []byte(`{}`), 0644)

	if err := profile.Snapshot(home, repo); err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(repo, ".claude.json")); err != nil {
		t.Error("Snapshot() did not copy .claude.json to repo")
	}
	if _, err := os.Stat(filepath.Join(repo, ".claude", "settings.json")); err != nil {
		t.Error("Snapshot() did not copy .claude/settings.json to repo")
	}
}

func TestRestore(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()

	os.WriteFile(filepath.Join(repo, ".claude.json"), []byte(`{"version":2}`), 0644)
	os.MkdirAll(filepath.Join(repo, ".claude"), 0755)
	os.WriteFile(filepath.Join(repo, ".claude", "settings.json"), []byte(`{"theme":"dark"}`), 0644)

	// Stale files in home that should be replaced
	os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"version":1}`), 0644)
	os.MkdirAll(filepath.Join(home, ".claude"), 0755)
	os.WriteFile(filepath.Join(home, ".claude", "old.json"), []byte(`old`), 0644)

	if err := profile.Restore(repo, home); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		t.Fatalf("Restore() did not write .claude.json: %v", err)
	}
	if string(data) != `{"version":2}` {
		t.Errorf(".claude.json content = %q, want %q", data, `{"version":2}`)
	}

	if _, err := os.Stat(filepath.Join(home, ".claude", "old.json")); !os.IsNotExist(err) {
		t.Error("Restore() left stale file old.json in ~/.claude/")
	}
}

func TestRestoreMissingSourceFiles(t *testing.T) {
	home := t.TempDir()
	repo := t.TempDir()
	if err := profile.Restore(repo, home); err != nil {
		t.Fatalf("Restore() with missing source files error = %v", err)
	}
}
