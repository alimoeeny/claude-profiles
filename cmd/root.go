package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ali/claude-profile-manager/internal/repopath"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "claude-profiles",
	Short: "Manage multiple Claude Code profiles",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// requireStore returns the store path and errors if it has not been initialised.
func requireStore() (string, error) {
	dir, err := repopath.Resolve()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(dir, "config.toml")); os.IsNotExist(err) {
		return "", fmt.Errorf("no profiles store found at %s — run 'claude-profiles init' to get started", dir)
	}
	return dir, nil
}
