package cmd

import (
	"fmt"

	"github.com/ali/claude-profile-switcher/internal/config"
	"github.com/ali/claude-profile-switcher/internal/git"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the active profile and its state",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	repoDir, err := requireRepo()
	if err != nil {
		return err
	}

	current, err := git.CurrentBranch(repoDir)
	if err != nil {
		return err
	}

	dirty, err := git.IsDirty(repoDir)
	if err != nil {
		return err
	}

	cfg, err := config.Load(repoDir)
	if err != nil {
		return err
	}

	state := "clean"
	if dirty {
		stat, _ := git.DiffStat(repoDir)
		state = fmt.Sprintf("dirty\n%s", stat)
	}

	remote := cfg.BackupRemote
	if remote == "" {
		remote = "(not configured)"
	}

	fmt.Printf("Profile: %s\nState:   %s\nRemote:  %s\n", current, state, remote)
	return nil
}
