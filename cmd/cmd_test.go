package cmd

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alimoeeny/claude-profiles/internal/claude"
	"github.com/alimoeeny/claude-profiles/internal/config"
	"github.com/alimoeeny/claude-profiles/internal/profile"
)

// setupStore creates an isolated store and home dir for testing.
//
// profiles maps profile name → content written to .claude.json inside that profile's zip.
// cfg is written to config.toml; pass nil for an empty config.
// Sets CLAUDE_PROFILES_DIR and HOME env vars for the duration of the test.
// Returns (storeDir, homeDir).
func setupStore(t *testing.T, cfg *config.Config, profiles map[string]string) (string, string) {
	t.Helper()

	storeDir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("CLAUDE_PROFILES_DIR", storeDir)
	t.Setenv("HOME", homeDir)

	if err := os.MkdirAll(filepath.Join(storeDir, "profiles"), 0755); err != nil {
		t.Fatalf("MkdirAll profiles: %v", err)
	}

	for name, claudeJSON := range profiles {
		// Snapshot a minimal temp home so each profile has a real zip.
		tmp := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmp, ".claude.json"), []byte(claudeJSON), 0644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		if err := profile.Snapshot(tmp, storeDir, name); err != nil {
			t.Fatalf("Snapshot %q: %v", name, err)
		}
	}

	if cfg == nil {
		cfg = &config.Config{}
	}
	if err := config.Save(storeDir, cfg); err != nil {
		t.Fatalf("config.Save: %v", err)
	}

	return storeDir, homeDir
}

// captureOutput replaces os.Stdout with a pipe for the duration of fn,
// then returns everything fn wrote.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	return buf.String()
}

// notRunning stubs claude.RunPgrep to report Claude is not running.
// Uses a real "sh -c exit 1" to produce a genuine *exec.ExitError with code 1,
// which IsRunning interprets as "not running, no error".
func notRunning(t *testing.T) {
	t.Helper()
	orig := claude.RunPgrep
	claude.RunPgrep = func(string) error {
		return exec.Command("sh", "-c", "exit 1").Run()
	}
	t.Cleanup(func() { claude.RunPgrep = orig })
}

// injectStdin sets the package-level stdin var and restores it after the test.
func injectStdin(t *testing.T, input string) {
	t.Helper()
	orig := stdin
	stdin = strings.NewReader(input)
	t.Cleanup(func() { stdin = orig })
}

// --- runList ---

func TestRunList_NoStore(t *testing.T) {
	t.Setenv("CLAUDE_PROFILES_DIR", t.TempDir()) // dir exists but no config.toml

	err := runList(nil, nil)
	if err == nil {
		t.Fatal("expected error when store not initialised")
	}
	if !strings.Contains(err.Error(), "init") {
		t.Errorf("error %q should mention 'init'", err.Error())
	}
}

func TestRunList_Empty(t *testing.T) {
	setupStore(t, &config.Config{}, nil)

	out := captureOutput(t, func() {
		if err := runList(nil, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if out != "" {
		t.Errorf("expected empty output for empty store, got %q", out)
	}
}

func TestRunList_ShowsCurrent(t *testing.T) {
	setupStore(t, &config.Config{Current: "alpha", SnapshotHash: ""}, map[string]string{
		"alpha": `{"profile":"alpha"}`,
		"beta":  `{"profile":"beta"}`,
	})
	// homeDir is empty → HashHome = "" = SnapshotHash → clean

	out := captureOutput(t, func() {
		if err := runList(nil, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "* alpha") {
		t.Errorf("output %q should contain '* alpha'", out)
	}
	if !strings.Contains(out, "  beta") {
		t.Errorf("output %q should contain '  beta'", out)
	}
}

func TestRunList_ShowsDirty(t *testing.T) {
	setupStore(t, &config.Config{Current: "alpha", SnapshotHash: "stale-hash"}, map[string]string{
		"alpha": `{}`,
	})
	// homeDir is empty → HashHome = "" ≠ "stale-hash" → dirty

	out := captureOutput(t, func() {
		if err := runList(nil, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "* alpha (dirty)") {
		t.Errorf("output %q should contain '* alpha (dirty)'", out)
	}
}

// --- runStatus ---

func TestRunStatus_Clean(t *testing.T) {
	setupStore(t, &config.Config{Current: "main", SnapshotHash: ""}, map[string]string{
		"main": `{}`,
	})

	out := captureOutput(t, func() {
		if err := runStatus(nil, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Profile: main") {
		t.Errorf("output %q should contain 'Profile: main'", out)
	}
	if !strings.Contains(out, "State:   clean") {
		t.Errorf("output %q should contain 'State:   clean'", out)
	}
	if !strings.Contains(out, "Remote:  (not configured)") {
		t.Errorf("output %q should contain 'Remote:  (not configured)'", out)
	}
}

func TestRunStatus_Dirty(t *testing.T) {
	setupStore(t, &config.Config{Current: "main", SnapshotHash: "stale"}, map[string]string{
		"main": `{}`,
	})

	out := captureOutput(t, func() {
		if err := runStatus(nil, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "dirty (unsaved changes)") {
		t.Errorf("output %q should contain 'dirty (unsaved changes)'", out)
	}
}

func TestRunStatus_WithRemote(t *testing.T) {
	setupStore(t, &config.Config{
		Current:      "main",
		BackupRemote: "user@host:/backups/",
	}, map[string]string{"main": `{}`})

	out := captureOutput(t, func() {
		if err := runStatus(nil, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "user@host:/backups/") {
		t.Errorf("output %q should contain remote path", out)
	}
}

// --- runSwitch ---

func TestRunSwitch_ProfileNotFound(t *testing.T) {
	notRunning(t)
	setupStore(t, &config.Config{Current: "alpha"}, map[string]string{"alpha": `{}`})

	err := runSwitch(nil, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error %q should contain 'not found'", err.Error())
	}
}

func TestRunSwitch_AlreadyOnProfile(t *testing.T) {
	notRunning(t)
	setupStore(t, &config.Config{Current: "alpha"}, map[string]string{"alpha": `{}`})

	out := captureOutput(t, func() {
		if err := runSwitch(nil, []string{"alpha"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Already on profile") {
		t.Errorf("output %q should say 'Already on profile'", out)
	}
}

func TestRunSwitch_CleanSwitch(t *testing.T) {
	notRunning(t)
	injectStdin(t, "") // clean home → no prompt
	storeDir, homeDir := setupStore(t,
		&config.Config{Current: "alpha", SnapshotHash: ""},
		map[string]string{
			"alpha": `{"profile":"alpha"}`,
			"beta":  `{"profile":"beta"}`,
		},
	)

	if err := runSwitch(nil, []string{"beta"}); err != nil {
		t.Fatalf("runSwitch error: %v", err)
	}

	// Home should now contain beta's .claude.json
	data, err := os.ReadFile(filepath.Join(homeDir, ".claude.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "beta") {
		t.Errorf(".claude.json = %q, want beta content", string(data))
	}

	// Config should be updated
	cfg, err := config.Load(storeDir)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.Current != "beta" {
		t.Errorf("config.Current = %q, want 'beta'", cfg.Current)
	}
}

func TestRunSwitch_Cancel(t *testing.T) {
	notRunning(t)
	injectStdin(t, "c\n")
	storeDir, _ := setupStore(t,
		&config.Config{Current: "alpha", SnapshotHash: "stale"},
		map[string]string{
			"alpha": `{"profile":"alpha"}`,
			"beta":  `{"profile":"beta"}`,
		},
	)

	out := captureOutput(t, func() {
		if err := runSwitch(nil, []string{"beta"}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Cancelled") {
		t.Errorf("output %q should say 'Cancelled'", out)
	}

	// Config must not have changed
	cfg, err := config.Load(storeDir)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.Current != "alpha" {
		t.Errorf("config.Current = %q, want 'alpha' after cancel", cfg.Current)
	}
}

// --- runDelete ---

func TestRunDelete_DefaultProfile(t *testing.T) {
	setupStore(t, &config.Config{Current: "main"}, map[string]string{
		"main":    `{}`,
		"default": `{}`,
	})

	err := runDelete(nil, []string{"default"})
	if err == nil {
		t.Fatal("expected error when deleting 'default'")
	}
	if !strings.Contains(err.Error(), "cannot delete 'default'") {
		t.Errorf("error %q should mention cannot delete default", err.Error())
	}
}

func TestRunDelete_ActiveProfile(t *testing.T) {
	setupStore(t, &config.Config{Current: "alpha"}, map[string]string{"alpha": `{}`})

	err := runDelete(nil, []string{"alpha"})
	if err == nil {
		t.Fatal("expected error when deleting active profile")
	}
	if !strings.Contains(err.Error(), "cannot delete active profile") {
		t.Errorf("error %q should mention cannot delete active", err.Error())
	}
}

func TestRunDelete_NotFound(t *testing.T) {
	setupStore(t, &config.Config{Current: "alpha"}, map[string]string{"alpha": `{}`})

	err := runDelete(nil, []string{"ghost"})
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error %q should say 'not found'", err.Error())
	}
}

func TestRunDelete_Cancelled(t *testing.T) {
	injectStdin(t, "n\n")
	storeDir, _ := setupStore(t,
		&config.Config{Current: "alpha"},
		map[string]string{"alpha": `{}`, "beta": `{}`},
	)

	if err := runDelete(nil, []string{"beta"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !profile.ProfileExists(storeDir, "beta") {
		t.Error("'beta' should still exist after cancel")
	}
}

func TestRunDelete_Confirmed(t *testing.T) {
	injectStdin(t, "y\n")
	storeDir, _ := setupStore(t,
		&config.Config{Current: "alpha"},
		map[string]string{"alpha": `{}`, "beta": `{}`},
	)

	if err := runDelete(nil, []string{"beta"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProfileExists(storeDir, "beta") {
		t.Error("'beta' should be deleted after confirmation")
	}
}

// --- runDuplicate ---

func TestRunDuplicate_AlreadyExists(t *testing.T) {
	notRunning(t)
	setupStore(t, &config.Config{Current: "alpha"}, map[string]string{
		"alpha": `{}`,
		"beta":  `{}`,
	})

	err := runDuplicate(nil, []string{"beta"})
	if err == nil {
		t.Fatal("expected error when duplicate name already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error %q should say 'already exists'", err.Error())
	}
}

func TestRunDuplicate_CleanCopy(t *testing.T) {
	notRunning(t)
	injectStdin(t, "") // clean home → no prompt
	storeDir, _ := setupStore(t,
		&config.Config{Current: "alpha", SnapshotHash: ""},
		map[string]string{"alpha": `{"profile":"alpha"}`},
	)

	if err := runDuplicate(nil, []string{"alpha-copy"}); err != nil {
		t.Fatalf("runDuplicate error: %v", err)
	}

	if !profile.ProfileExists(storeDir, "alpha-copy") {
		t.Error("'alpha-copy' profile should exist after duplicate")
	}

	cfg, err := config.Load(storeDir)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.Current != "alpha-copy" {
		t.Errorf("config.Current = %q, want 'alpha-copy'", cfg.Current)
	}
}

// --- zipDir ---

func TestZipDir_ContainsAllFiles(t *testing.T) {
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("aaa"), 0644)
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("bbb"), 0644)

	dest := filepath.Join(t.TempDir(), "out.zip")
	if err := zipDir(src, dest); err != nil {
		t.Fatalf("zipDir error: %v", err)
	}

	zr, err := openZipIndex(t, dest)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	want := map[string]string{"a.txt": "aaa", "sub/b.txt": "bbb"}
	for name, wantContent := range want {
		got, ok := zr[name]
		if !ok {
			t.Errorf("zip missing entry %q", name)
			continue
		}
		if got != wantContent {
			t.Errorf("zip[%q] = %q, want %q", name, got, wantContent)
		}
	}
}

func TestZipDir_EmptyDir(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(t.TempDir(), "out.zip")

	if err := zipDir(src, dest); err != nil {
		t.Fatalf("zipDir on empty dir error: %v", err)
	}

	zr, err := openZipIndex(t, dest)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	if len(zr) != 0 {
		t.Errorf("expected empty zip, got %d entries", len(zr))
	}
}

// --- writeDefaultZip ---

func TestWriteDefaultZip_ContainsExpectedFiles(t *testing.T) {
	storeDir := t.TempDir()

	if err := writeDefaultZip(storeDir); err != nil {
		t.Fatalf("writeDefaultZip error: %v", err)
	}

	zipPath := filepath.Join(storeDir, "profiles", "default.zip")
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("default.zip not created: %v", err)
	}

	entries, err := openZipIndex(t, zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	if _, ok := entries[".claude.json"]; !ok {
		t.Errorf("default.zip missing .claude.json; entries: %v", keys(entries))
	}
}

// openZipIndex opens a zip file and returns a map of entry name → content.
func openZipIndex(t *testing.T, path string) (map[string]string, error) {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	out := make(map[string]string)
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		io.Copy(&buf, rc)
		rc.Close()
		out[f.Name] = buf.String()
	}
	return out, nil
}

// keys returns the keys of a map for use in error messages.
func keys(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

// --- runCreateFresh ---

func TestRunCreateFresh_AlreadyExists(t *testing.T) {
	notRunning(t)
	setupStore(t, &config.Config{Current: "main"}, map[string]string{
		"main":    `{}`,
		"default": `{}`,
		"work":    `{}`,
	})

	err := runCreateFresh(nil, []string{"work"})
	if err == nil {
		t.Fatal("expected error when name already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error %q should say 'already exists'", err.Error())
	}
}

func TestRunCreateFresh_NoDefaultProfile(t *testing.T) {
	notRunning(t)
	setupStore(t, &config.Config{Current: "main"}, map[string]string{"main": `{}`})
	// No "default" profile in the store

	err := runCreateFresh(nil, []string{"newprofile"})
	if err == nil {
		t.Fatal("expected error when 'default' profile missing")
	}
	if !strings.Contains(err.Error(), "'default' profile not found") {
		t.Errorf("error %q should mention default profile not found", err.Error())
	}
}

func TestRunCreateFresh_Success(t *testing.T) {
	notRunning(t)
	injectStdin(t, "") // clean home → no prompt
	storeDir, homeDir := setupStore(t,
		&config.Config{Current: "main", SnapshotHash: ""},
		map[string]string{
			"main":    `{"profile":"main"}`,
			"default": `{"profile":"default"}`,
		},
	)

	if err := runCreateFresh(nil, []string{"fresh"}); err != nil {
		t.Fatalf("runCreateFresh error: %v", err)
	}

	// New profile zip must exist
	if !profile.ProfileExists(storeDir, "fresh") {
		t.Error("'fresh' profile should exist after create_fresh")
	}

	// Config must reflect the new active profile
	cfg, err := config.Load(storeDir)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.Current != "fresh" {
		t.Errorf("config.Current = %q, want 'fresh'", cfg.Current)
	}

	// Home should now contain the default profile's content
	data, err := os.ReadFile(filepath.Join(homeDir, ".claude.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "default") {
		t.Errorf(".claude.json = %q, want default content", string(data))
	}
}
