package profile_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alimoeeny/claude-profiles/internal/profile"
)

func makeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"version":1}`), 0644)
	os.MkdirAll(filepath.Join(home, ".claude", "mcp"), 0755)
	os.WriteFile(filepath.Join(home, ".claude", "settings.json"), []byte(`{}`), 0644)
	return home
}

func TestSnapshot(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	if err := profile.Snapshot(home, store, "test"); err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if !profile.ProfileExists(store, "test") {
		t.Error("Snapshot() did not create test.zip")
	}
}

func TestRestore(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	if err := profile.Snapshot(home, store, "test"); err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	// Overwrite home with different content
	home2 := t.TempDir()
	os.WriteFile(filepath.Join(home2, ".claude.json"), []byte(`{"version":99}`), 0644)

	if err := profile.Restore(store, home2, "test"); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(home2, ".claude.json"))
	if err != nil {
		t.Fatalf("Restore() did not write .claude.json: %v", err)
	}
	if string(data) != `{"version":1}` {
		t.Errorf(".claude.json = %q, want %q", data, `{"version":1}`)
	}

	if _, err := os.Stat(filepath.Join(home2, ".claude", "settings.json")); err != nil {
		t.Error("Restore() did not restore .claude/settings.json")
	}
}

func TestRestoreRemovesStaleFiles(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	if err := profile.Snapshot(home, store, "test"); err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	// Add a stale file to home
	os.WriteFile(filepath.Join(home, ".claude", "stale.json"), []byte(`old`), 0644)

	if err := profile.Restore(store, home, "test"); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, ".claude", "stale.json")); !os.IsNotExist(err) {
		t.Error("Restore() left stale file stale.json")
	}
}

func TestRestoreMissingZip(t *testing.T) {
	home := t.TempDir()
	store := t.TempDir()
	if err := profile.Restore(store, home, "nonexistent"); err != nil {
		t.Fatalf("Restore() with missing zip error = %v", err)
	}
}

func TestListProfiles(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	profile.Snapshot(home, store, "alpha")
	profile.Snapshot(home, store, "beta")

	names, err := profile.ListProfiles(store)
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	if len(names) != 2 || names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("ListProfiles() = %v, want [alpha beta]", names)
	}
}

func TestProfileExists(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	profile.Snapshot(home, store, "myprofile")

	if !profile.ProfileExists(store, "myprofile") {
		t.Error("ProfileExists() = false for existing profile")
	}
	if profile.ProfileExists(store, "missing") {
		t.Error("ProfileExists() = true for missing profile")
	}
}

func TestDeleteProfile(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	profile.Snapshot(home, store, "todelete")

	if err := profile.DeleteProfile(store, "todelete"); err != nil {
		t.Fatalf("DeleteProfile() error = %v", err)
	}
	if profile.ProfileExists(store, "todelete") {
		t.Error("profile still exists after DeleteProfile()")
	}
}

func TestDuplicateProfile(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	profile.Snapshot(home, store, "original")

	if err := profile.DuplicateProfile(store, "original", "copy"); err != nil {
		t.Fatalf("DuplicateProfile() error = %v", err)
	}
	if !profile.ProfileExists(store, "copy") {
		t.Error("DuplicateProfile() did not create copy")
	}

	// Restore from copy and verify contents match
	home2 := t.TempDir()
	if err := profile.Restore(store, home2, "copy"); err != nil {
		t.Fatalf("Restore() from copy error = %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(home2, ".claude.json"))
	if string(data) != `{"version":1}` {
		t.Errorf("copy .claude.json = %q, want %q", data, `{"version":1}`)
	}
}

func TestHashHome(t *testing.T) {
	home := makeHome(t)

	h1, err := profile.HashHome(home)
	if err != nil {
		t.Fatalf("HashHome() error = %v", err)
	}
	if h1 == "" {
		t.Error("HashHome() returned empty string for non-empty home")
	}

	// Modify a file — hash must change
	os.WriteFile(filepath.Join(home, ".claude", "settings.json"), []byte(`{"theme":"dark"}`), 0644)
	h2, _ := profile.HashHome(home)
	if h1 == h2 {
		t.Error("HashHome() same hash after file change")
	}
}

func TestIsDirty(t *testing.T) {
	home := makeHome(t)

	storedHash, _ := profile.HashHome(home)

	dirty, err := profile.IsDirty(home, storedHash)
	if err != nil {
		t.Fatalf("IsDirty() error = %v", err)
	}
	if dirty {
		t.Error("IsDirty() = true when hash matches, want false")
	}

	os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"changed":true}`), 0644)
	dirty, _ = profile.IsDirty(home, storedHash)
	if !dirty {
		t.Error("IsDirty() = false after file change, want true")
	}
}
