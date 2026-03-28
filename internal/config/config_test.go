package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alimoeeny/claude-profiles/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BackupRemote != "" {
		t.Errorf("BackupRemote = %q, want empty", cfg.BackupRemote)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{BackupRemote: "git@github.com:user/profiles.git"}

	if err := config.Save(dir, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}
	if loaded.BackupRemote != cfg.BackupRemote {
		t.Errorf("BackupRemote = %q, want %q", loaded.BackupRemote, cfg.BackupRemote)
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	os.Remove(filepath.Join(dir, "config.toml"))

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load() with missing file error = %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
}

func TestSaveAndLoad_AllFields(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		Current:      "work",
		BackupRemote: "user@host:/backups/",
		SnapshotHash: "abc123",
	}

	if err := config.Save(dir, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Current != cfg.Current {
		t.Errorf("Current = %q, want %q", loaded.Current, cfg.Current)
	}
	if loaded.BackupRemote != cfg.BackupRemote {
		t.Errorf("BackupRemote = %q, want %q", loaded.BackupRemote, cfg.BackupRemote)
	}
	if loaded.SnapshotHash != cfg.SnapshotHash {
		t.Errorf("SnapshotHash = %q, want %q", loaded.SnapshotHash, cfg.SnapshotHash)
	}
}

func TestLoad_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.toml"), []byte("current = [broken"), 0644)

	_, err := config.Load(dir)
	if err == nil {
		t.Fatal("Load() expected error for corrupt TOML, got nil")
	}
}

func TestSave_DirectoryMissing(t *testing.T) {
	err := config.Save("/nonexistent/dir/that/does/not/exist", &config.Config{})
	if err == nil {
		t.Fatal("Save() expected error for missing directory, got nil")
	}
}

func TestSaveAndLoad_EmptySnapshotHash(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{Current: "main", SnapshotHash: ""}

	if err := config.Save(dir, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := config.Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.SnapshotHash != "" {
		t.Errorf("SnapshotHash = %q, want empty string", loaded.SnapshotHash)
	}
}
