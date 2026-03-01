package cmd

import (
	"fmt"
	"os"

	"github.com/ali/claude-profile-switcher/internal/claude"
	"github.com/ali/claude-profile-switcher/internal/config"
	"github.com/ali/claude-profile-switcher/internal/git"
	"github.com/ali/claude-profile-switcher/internal/profile"
	"github.com/ali/claude-profile-switcher/internal/prompt"
	"github.com/spf13/cobra"
)

var duplicateCmd = &cobra.Command{
	Use:   "duplicate <name>",
	Short: "Duplicate the current profile into a new one and switch to it",
	Args:  cobra.ExactArgs(1),
	RunE:  runDuplicate,
}

func init() {
	rootCmd.AddCommand(duplicateCmd)
}

func runDuplicate(cmd *cobra.Command, args []string) error {
	name := args[0]

	repoDir, err := requireRepo()
	if err != nil {
		return err
	}

	running, err := claude.IsRunning()
	if err != nil {
		return err
	}
	if running {
		return fmt.Errorf("Claude is running — close it before duplicating profiles")
	}

	if git.BranchExists(repoDir, name) {
		return fmt.Errorf("profile %q already exists", name)
	}

	action, err := prompt.HandleDirty(repoDir, true)
	if err != nil {
		return err
	}
	if action == prompt.ActionCancel {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := git.CreateBranch(repoDir, name); err != nil {
		return err
	}

	// If carrying over, snapshot the current home state and commit on the new branch
	if action == prompt.ActionCarryOver {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		if err := profile.Snapshot(home, repoDir); err != nil {
			return err
		}
		cfg, _ := config.Load(repoDir)
		if err := config.Save(repoDir, cfg); err != nil {
			return err
		}
		if err := git.CommitAll(repoDir, "duplicate: carry over changes from previous profile"); err != nil {
			return err
		}
	}

	fmt.Printf("Created and switched to profile %q\n", name)
	return nil
}
