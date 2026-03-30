package profile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alimoeeny/claude-profiles/internal/profile"
)

func TestDiffSnapshot_CleanProfile(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	if err := profile.Snapshot(home, store, "main"); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	result, err := profile.DiffSnapshot(home, store, "main")
	if err != nil {
		t.Fatalf("DiffSnapshot: %v", err)
	}

	if result.Added != 0 || result.Removed != 0 || result.Modified != 0 {
		t.Errorf("expected clean profile, got added=%d removed=%d modified=%d",
			result.Added, result.Removed, result.Modified)
	}
}

func TestDiffSnapshot_ModifiedFile(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	if err := profile.Snapshot(home, store, "main"); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	// Modify .claude.json after snapshot
	os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"version":2}`), 0644)

	result, err := profile.DiffSnapshot(home, store, "main")
	if err != nil {
		t.Fatalf("DiffSnapshot: %v", err)
	}

	if result.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", result.Modified)
	}
	if result.Entries[0].Path != ".claude.json" {
		t.Errorf("expected modified path .claude.json, got %q", result.Entries[0].Path)
	}
	if result.Entries[0].Status != profile.DiffModified {
		t.Errorf("expected DiffModified status")
	}
}

func TestDiffSnapshot_AddedFile(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	if err := profile.Snapshot(home, store, "main"); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	// Add a new file after snapshot
	os.WriteFile(filepath.Join(home, ".claude", "new.json"), []byte(`{}`), 0644)

	result, err := profile.DiffSnapshot(home, store, "main")
	if err != nil {
		t.Fatalf("DiffSnapshot: %v", err)
	}

	if result.Added != 1 {
		t.Errorf("expected 1 added, got %d", result.Added)
	}

	found := false
	for _, e := range result.Entries {
		if e.Path == ".claude/new.json" && e.Status == profile.DiffAdded {
			found = true
		}
	}
	if !found {
		t.Error("expected DiffAdded entry for .claude/new.json")
	}
}

func TestDiffSnapshot_RemovedFile(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	if err := profile.Snapshot(home, store, "main"); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	// Remove a file that was in the snapshot
	os.Remove(filepath.Join(home, ".claude", "settings.json"))

	result, err := profile.DiffSnapshot(home, store, "main")
	if err != nil {
		t.Fatalf("DiffSnapshot: %v", err)
	}

	if result.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", result.Removed)
	}

	found := false
	for _, e := range result.Entries {
		if e.Path == ".claude/settings.json" && e.Status == profile.DiffRemoved {
			found = true
		}
	}
	if !found {
		t.Error("expected DiffRemoved entry for .claude/settings.json")
	}
}

func TestDiffSnapshot_MultipleChanges(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	// Snapshot has: .claude.json, .claude/settings.json
	if err := profile.Snapshot(home, store, "main"); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	// Modify .claude.json
	os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"version":99}`), 0644)
	// Remove settings.json
	os.Remove(filepath.Join(home, ".claude", "settings.json"))
	// Add a new file
	os.WriteFile(filepath.Join(home, ".claude", "extra.json"), []byte(`{}`), 0644)

	result, err := profile.DiffSnapshot(home, store, "main")
	if err != nil {
		t.Fatalf("DiffSnapshot: %v", err)
	}

	if result.Added != 1 {
		t.Errorf("expected Added=1, got %d", result.Added)
	}
	if result.Removed != 1 {
		t.Errorf("expected Removed=1, got %d", result.Removed)
	}
	if result.Modified != 1 {
		t.Errorf("expected Modified=1, got %d", result.Modified)
	}
}

func TestDiffSnapshot_EmptyProfile(t *testing.T) {
	home := t.TempDir() // no .claude.json, no .claude/
	store := t.TempDir()

	if err := profile.Snapshot(home, store, "empty"); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	result, err := profile.DiffSnapshot(home, store, "empty")
	if err != nil {
		t.Fatalf("DiffSnapshot: %v", err)
	}

	if result.Added != 0 || result.Removed != 0 || result.Modified != 0 {
		t.Errorf("expected all zeros for empty profile, got %+v", result)
	}
}

func TestDiffSnapshot_MissingZip(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	_, err := profile.DiffSnapshot(home, store, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing snapshot zip, got nil")
	}
}
