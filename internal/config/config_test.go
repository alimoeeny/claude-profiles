package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ali/claude-profile-switcher/internal/config"
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
