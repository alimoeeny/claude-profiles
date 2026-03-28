package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const filename = "config.toml"

type Config struct {
	Current      string `toml:"current"`
	BackupRemote string `toml:"backup_remote"`
	SnapshotHash string `toml:"snapshot_hash"`
}

// Load reads config.toml from repoDir. Returns empty config if file doesn't exist.
func Load(repoDir string) (*Config, error) {
	cfg := &Config{}
	path := filepath.Join(repoDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("config: decode %s: %w", path, err)
	}
	return cfg, nil
}

// Save writes cfg to config.toml in repoDir.
func Save(repoDir string, cfg *Config) error {
	path := filepath.Join(repoDir, filename)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("config: create %s: %w", path, err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}
