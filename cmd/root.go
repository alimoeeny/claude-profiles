package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alimoeeny/claude-profiles/internal/repopath"
	"github.com/spf13/cobra"
)

// stdin is the reader used for all interactive prompts.
// Tests replace it with a strings.Reader to inject input without touching os.Stdin.
var stdin io.Reader = os.Stdin

// Version is overridden at link time by GoReleaser via -ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:          "claude-profiles",
	Short:        "Manage multiple Claude Code profiles",
	Version:      Version,
	SilenceUsage: true,
	RunE:         runList,
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
