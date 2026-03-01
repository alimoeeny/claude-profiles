package cmd

import (
	"fmt"
	"os"

	"github.com/ali/claude-profile-switcher/internal/git"
	"github.com/ali/claude-profile-switcher/internal/repopath"
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

// requireRepo returns the repo path and errors if it's not initialised.
func requireRepo() (string, error) {
	dir, err := repopath.Resolve()
	if err != nil {
		return "", err
	}
	if !git.IsRepo(dir) {
		return "", fmt.Errorf("no profiles repo found at %s — run 'claude-profiles init' to get started", dir)
	}
	return dir, nil
}
