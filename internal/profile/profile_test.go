package profile_test

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
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

// --- HashHome edge cases ---

func TestHashHome_EmptyHome(t *testing.T) {
	home := t.TempDir()
	h, err := profile.HashHome(home)
	if err != nil {
		t.Fatalf("HashHome() error = %v", err)
	}
	if h != "" {
		t.Errorf("HashHome() = %q on empty home, want empty string", h)
	}
}

func TestHashHome_Determinism(t *testing.T) {
	home := makeHome(t)
	h1, err := profile.HashHome(home)
	if err != nil {
		t.Fatalf("first HashHome() error = %v", err)
	}
	h2, err := profile.HashHome(home)
	if err != nil {
		t.Fatalf("second HashHome() error = %v", err)
	}
	if h1 != h2 {
		t.Errorf("HashHome() not deterministic: %q != %q", h1, h2)
	}
}

func TestHashHome_OnlyJsonFile(t *testing.T) {
	home := t.TempDir()
	os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{}`), 0644)

	h, err := profile.HashHome(home)
	if err != nil {
		t.Fatalf("HashHome() error = %v", err)
	}
	if h == "" {
		t.Error("HashHome() returned empty string when .claude.json present")
	}
}

func TestHashHome_OnlyDotClaudeDir(t *testing.T) {
	home := t.TempDir()
	os.MkdirAll(filepath.Join(home, ".claude"), 0755)
	os.WriteFile(filepath.Join(home, ".claude", "settings.json"), []byte(`{}`), 0644)

	h, err := profile.HashHome(home)
	if err != nil {
		t.Fatalf("HashHome() error = %v", err)
	}
	if h == "" {
		t.Error("HashHome() returned empty string when .claude/ dir present")
	}
}

func TestHashHome_SkipsGitDir(t *testing.T) {
	home := makeHome(t)
	gitDir := filepath.Join(home, ".claude", ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main"), 0644)

	h1, err := profile.HashHome(home)
	if err != nil {
		t.Fatalf("HashHome() error = %v", err)
	}

	// Modifying files inside .git must not change the hash.
	os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/other"), 0644)

	h2, err := profile.HashHome(home)
	if err != nil {
		t.Fatalf("HashHome() after .git change error = %v", err)
	}
	if h1 != h2 {
		t.Error("HashHome() changed after modifying .git — .git dir should be skipped")
	}
}

func TestHashHome_SkipsSymlinks(t *testing.T) {
	home := makeHome(t)
	target := filepath.Join(home, "real_file.txt")
	os.WriteFile(target, []byte("real"), 0644)
	os.Symlink(target, filepath.Join(home, ".claude", "link"))

	_, err := profile.HashHome(home)
	if err != nil {
		t.Fatalf("HashHome() error on symlink = %v", err)
	}
}

// --- Snapshot / Restore edge cases ---

func TestSnapshot_OnlyJsonNoDir(t *testing.T) {
	home := t.TempDir()
	os.WriteFile(filepath.Join(home, ".claude.json"), []byte(`{"v":1}`), 0644)

	store := t.TempDir()
	if err := profile.Snapshot(home, store, "p"); err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	home2 := t.TempDir()
	if err := profile.Restore(store, home2, "p"); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	data, err := os.ReadFile(filepath.Join(home2, ".claude.json"))
	if err != nil {
		t.Fatalf(".claude.json missing after Restore: %v", err)
	}
	if string(data) != `{"v":1}` {
		t.Errorf(".claude.json = %q, want %q", data, `{"v":1}`)
	}
	if _, err := os.Stat(filepath.Join(home2, ".claude")); !os.IsNotExist(err) {
		t.Error(".claude/ dir should not exist when not in snapshot")
	}
}

func TestSnapshot_OnlyDirNoJson(t *testing.T) {
	home := t.TempDir()
	os.MkdirAll(filepath.Join(home, ".claude"), 0755)
	os.WriteFile(filepath.Join(home, ".claude", "settings.json"), []byte(`{}`), 0644)

	store := t.TempDir()
	if err := profile.Snapshot(home, store, "p"); err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	home2 := t.TempDir()
	if err := profile.Restore(store, home2, "p"); err != nil {
		t.Fatalf("Restore() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(home2, ".claude", "settings.json")); err != nil {
		t.Error(".claude/settings.json missing after Restore")
	}
	if _, err := os.Stat(filepath.Join(home2, ".claude.json")); !os.IsNotExist(err) {
		t.Error(".claude.json should not exist when not in snapshot")
	}
}

func TestRestore_ZipPathTraversal(t *testing.T) {
	homeDir := t.TempDir()
	storeDir := t.TempDir()

	// Manually craft a zip with a path-traversal entry name.
	profilesDir := filepath.Join(storeDir, "profiles")
	os.MkdirAll(profilesDir, 0755)
	zp := filepath.Join(profilesDir, "evil.zip")
	f, err := os.Create(zp)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	w := zip.NewWriter(f)
	entry, err := w.Create("../../evil.txt")
	if err != nil {
		t.Fatalf("zip entry: %v", err)
	}
	entry.Write([]byte("pwned"))
	w.Close()
	f.Close()

	// Restore should reject the traversal entry (return an error).
	err = profile.Restore(storeDir, homeDir, "evil")
	if err == nil {
		t.Fatal("Restore() should have returned an error for zip path traversal")
	}
	if !strings.Contains(err.Error(), "unsafe path") {
		t.Errorf("error %q does not mention unsafe path", err.Error())
	}

	// The traversal target must not have been created.
	evilTarget := filepath.Clean(filepath.Join(homeDir, "../../evil.txt"))
	if _, statErr := os.Stat(evilTarget); statErr == nil {
		t.Fatal("Restore() wrote a file outside homeDir — path traversal not prevented")
	}
}

// --- Profile management edge cases ---

func TestDeleteProfile_NonExistent(t *testing.T) {
	store := t.TempDir()
	os.MkdirAll(filepath.Join(store, "profiles"), 0755)

	err := profile.DeleteProfile(store, "ghost")
	if err == nil {
		t.Fatal("DeleteProfile() expected error for nonexistent profile")
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("error %q does not mention profile name", err.Error())
	}
}

func TestDuplicateProfile_SrcMissing(t *testing.T) {
	store := t.TempDir()
	os.MkdirAll(filepath.Join(store, "profiles"), 0755)

	err := profile.DuplicateProfile(store, "missing", "copy")
	if err == nil {
		t.Fatal("DuplicateProfile() expected error when src does not exist")
	}
}

// --- ListProfiles edge cases ---

func TestListProfiles_EmptyDir(t *testing.T) {
	store := t.TempDir()
	os.MkdirAll(filepath.Join(store, "profiles"), 0755)

	names, err := profile.ListProfiles(store)
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	if len(names) != 0 {
		t.Errorf("ListProfiles() = %v, want empty slice", names)
	}
}

func TestListProfiles_DirMissing(t *testing.T) {
	store := t.TempDir()
	// No profiles/ subdir created.

	_, err := profile.ListProfiles(store)
	if err == nil {
		t.Fatal("ListProfiles() expected error when profiles dir is missing")
	}
}

func TestListProfiles_IgnoresNonZip(t *testing.T) {
	home := makeHome(t)
	store := t.TempDir()

	profile.Snapshot(home, store, "alpha")
	os.WriteFile(filepath.Join(store, "profiles", "readme.txt"), []byte("ignore me"), 0644)

	names, err := profile.ListProfiles(store)
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}
	if len(names) != 1 || names[0] != "alpha" {
		t.Errorf("ListProfiles() = %v, want [alpha]", names)
	}
}
