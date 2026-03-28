package prompt_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alimoeeny/claude-profiles/internal/config"
	"github.com/alimoeeny/claude-profiles/internal/profile"
	"github.com/alimoeeny/claude-profiles/internal/prompt"
)

// --- Ask ---

func TestAsk_ReturnsInput(t *testing.T) {
	got, err := prompt.Ask(strings.NewReader("hello\n"), "prompt: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestAsk_TrimsWhitespace(t *testing.T) {
	got, err := prompt.Ask(strings.NewReader("  hello  \n"), "prompt: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestAsk_EOFReturnsPartialLine(t *testing.T) {
	// No trailing newline — EOF mid-line should return the partial string without error.
	got, err := prompt.Ask(strings.NewReader("partial"), "prompt: ")
	if err != nil {
		t.Fatalf("unexpected error on EOF: %v", err)
	}
	if got != "partial" {
		t.Fatalf("got %q, want %q", got, "partial")
	}
}

// --- Confirm ---

func TestConfirm_YesLower(t *testing.T) {
	ok, err := prompt.Confirm(strings.NewReader("y\n"), "confirm? ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected true for 'y'")
	}
}

func TestConfirm_YesUpper(t *testing.T) {
	ok, err := prompt.Confirm(strings.NewReader("Y\n"), "confirm? ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected true for 'Y'")
	}
}

func TestConfirm_No(t *testing.T) {
	ok, err := prompt.Confirm(strings.NewReader("n\n"), "confirm? ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected false for 'n'")
	}
}

func TestConfirm_Empty(t *testing.T) {
	ok, err := prompt.Confirm(strings.NewReader("\n"), "confirm? ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected false for empty input")
	}
}

// --- HandleDirty helpers ---

// setupDirtyState creates:
//   - a homeDir with .claude.json containing originalContent
//   - a storeDir with profiles/<profileName>.zip snapshotted from that home
//   - config.toml with SnapshotHash matching the original home
//
// It then overwrites .claude.json with modifiedContent so the home is dirty.
// Returns homeDir and storeDir.
func setupDirtyState(t *testing.T, profileName, originalContent, modifiedContent string) (homeDir, storeDir string) {
	t.Helper()

	homeDir = t.TempDir()
	storeDir = t.TempDir()

	// Write original content and snapshot it.
	writeHomeFile(t, homeDir, ".claude.json", originalContent)
	if err := profile.Snapshot(homeDir, storeDir, profileName); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	// Record the hash of the original state.
	originalHash, err := profile.HashHome(homeDir)
	if err != nil {
		t.Fatalf("HashHome: %v", err)
	}

	// Save config with the original hash.
	cfg := &config.Config{Current: profileName, SnapshotHash: originalHash}
	if err := config.Save(storeDir, cfg); err != nil {
		t.Fatalf("config.Save: %v", err)
	}

	// Now modify the home to make it dirty.
	writeHomeFile(t, homeDir, ".claude.json", modifiedContent)
	return homeDir, storeDir
}

// setupCleanState creates a homeDir and storeDir where the hash matches.
func setupCleanState(t *testing.T, profileName, content string) (homeDir, storeDir string) {
	t.Helper()

	homeDir = t.TempDir()
	storeDir = t.TempDir()

	writeHomeFile(t, homeDir, ".claude.json", content)
	if err := profile.Snapshot(homeDir, storeDir, profileName); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}

	hash, err := profile.HashHome(homeDir)
	if err != nil {
		t.Fatalf("HashHome: %v", err)
	}

	cfg := &config.Config{Current: profileName, SnapshotHash: hash}
	if err := config.Save(storeDir, cfg); err != nil {
		t.Fatalf("config.Save: %v", err)
	}
	return homeDir, storeDir
}

func writeHomeFile(t *testing.T, homeDir, name, content string) {
	t.Helper()
	path := filepath.Join(homeDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// --- HandleDirty ---

func TestHandleDirty_Clean(t *testing.T) {
	homeDir, storeDir := setupCleanState(t, "myprofile", `{"version":1}`)

	// Pass a reader that panics if touched — proves stdin is never read on clean state.
	action, err := prompt.HandleDirty(panicReader{t}, homeDir, storeDir, "myprofile", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != prompt.ActionSave {
		t.Fatalf("expected ActionSave, got %v", action)
	}
}

func TestHandleDirty_DirtySave(t *testing.T) {
	homeDir, storeDir := setupDirtyState(t, "myprofile", `{"version":1}`, `{"version":2}`)

	action, err := prompt.HandleDirty(strings.NewReader("s\n"), homeDir, storeDir, "myprofile", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != prompt.ActionSave {
		t.Fatalf("expected ActionSave, got %v", action)
	}

	// Config hash should now match the current (modified) home.
	currentHash, err := profile.HashHome(homeDir)
	if err != nil {
		t.Fatalf("HashHome: %v", err)
	}
	cfg, err := config.Load(storeDir)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.SnapshotHash != currentHash {
		t.Fatalf("SnapshotHash not updated: got %q, want %q", cfg.SnapshotHash, currentHash)
	}
}

func TestHandleDirty_DirtyDiscard(t *testing.T) {
	homeDir, storeDir := setupDirtyState(t, "myprofile", `{"version":1}`, `{"version":2}`)

	action, err := prompt.HandleDirty(strings.NewReader("d\n"), homeDir, storeDir, "myprofile", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != prompt.ActionDiscard {
		t.Fatalf("expected ActionDiscard, got %v", action)
	}

	// Home should be restored to the original content.
	data, err := os.ReadFile(filepath.Join(homeDir, ".claude.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != `{"version":1}` {
		t.Fatalf("home not restored: got %q, want %q", string(data), `{"version":1}`)
	}
}

func TestHandleDirty_DirtyCancel(t *testing.T) {
	homeDir, storeDir := setupDirtyState(t, "myprofile", `{"version":1}`, `{"version":2}`)

	action, err := prompt.HandleDirty(strings.NewReader("c\n"), homeDir, storeDir, "myprofile", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != prompt.ActionCancel {
		t.Fatalf("expected ActionCancel, got %v", action)
	}

	// Home should still have the modified content — nothing was changed.
	data, err := os.ReadFile(filepath.Join(homeDir, ".claude.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != `{"version":2}` {
		t.Fatalf("home was unexpectedly modified: got %q", string(data))
	}
}

func TestHandleDirty_DirtyInvalidThenSave(t *testing.T) {
	homeDir, storeDir := setupDirtyState(t, "myprofile", `{"version":1}`, `{"version":2}`)

	// "x" is invalid, retry loop should re-prompt; "s" then saves.
	action, err := prompt.HandleDirty(strings.NewReader("x\ns\n"), homeDir, storeDir, "myprofile", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != prompt.ActionSave {
		t.Fatalf("expected ActionSave after retry, got %v", action)
	}
}

func TestHandleDirty_AllowCarryOver_CarryOver(t *testing.T) {
	homeDir, storeDir := setupDirtyState(t, "myprofile", `{"version":1}`, `{"version":2}`)

	action, err := prompt.HandleDirty(strings.NewReader("c\n"), homeDir, storeDir, "myprofile", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != prompt.ActionCarryOver {
		t.Fatalf("expected ActionCarryOver, got %v", action)
	}
}

func TestHandleDirty_AllowCarryOver_Abort(t *testing.T) {
	homeDir, storeDir := setupDirtyState(t, "myprofile", `{"version":1}`, `{"version":2}`)

	action, err := prompt.HandleDirty(strings.NewReader("a\n"), homeDir, storeDir, "myprofile", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != prompt.ActionCancel {
		t.Fatalf("expected ActionCancel, got %v", action)
	}
}

// --- UpdateSnapshotHash ---

func TestUpdateSnapshotHash(t *testing.T) {
	homeDir, storeDir := setupCleanState(t, "myprofile", `{"version":1}`)

	// Modify the home and update the hash.
	writeHomeFile(t, homeDir, ".claude.json", `{"version":2}`)
	if err := prompt.UpdateSnapshotHash(homeDir, storeDir); err != nil {
		t.Fatalf("UpdateSnapshotHash: %v", err)
	}

	newHash, err := profile.HashHome(homeDir)
	if err != nil {
		t.Fatalf("HashHome: %v", err)
	}
	cfg, err := config.Load(storeDir)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.SnapshotHash != newHash {
		t.Fatalf("SnapshotHash not updated: got %q, want %q", cfg.SnapshotHash, newHash)
	}
}

// panicReader panics if Read is ever called, proving stdin is not touched.
type panicReader struct{ t *testing.T }

func (p panicReader) Read(_ []byte) (int, error) {
	p.t.Fatal("Read called unexpectedly — stdin should not be read for a clean home")
	return 0, nil
}
