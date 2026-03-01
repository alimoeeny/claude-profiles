package cmd

import (
	"fmt"
	"os"

	"github.com/ali/claude-profile-switcher/internal/claude"
	"github.com/ali/claude-profile-switcher/internal/git"
	"github.com/ali/claude-profile-switcher/internal/profile"
	"github.com/ali/claude-profile-switcher/internal/prompt"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch <profile>",
	Short: "Switch to a different profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runSwitch,
}

func init() {
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	target := args[0]

	repoDir, err := requireRepo()
	if err != nil {
		return err
	}

	running, err := claude.IsRunning()
	if err != nil {
		return err
	}
	if running {
		return fmt.Errorf("Claude is running — close it before switching profiles")
	}

	if !git.BranchExists(repoDir, target) {
		return fmt.Errorf("profile %q not found — run 'claude-profiles list' to see available profiles", target)
	}

	current, err := git.CurrentBranch(repoDir)
	if err != nil {
		return err
	}
	if current == target {
		fmt.Printf("Already on profile %q\n", target)
		return nil
	}

	action, err := prompt.HandleDirty(repoDir, false)
	if err != nil {
		return err
	}
	if action == prompt.ActionCancel {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := git.Checkout(repoDir, target); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := profile.Restore(repoDir, home); err != nil {
		return err
	}

	fmt.Printf("Switched to profile %q\n", target)
	return nil
}
