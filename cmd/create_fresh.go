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

var createFreshCmd = &cobra.Command{
	Use:   "create_fresh <name>",
	Short: "Create a new profile from the default branch",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateFresh,
}

func init() {
	rootCmd.AddCommand(createFreshCmd)
}

func runCreateFresh(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("Claude is running — close it before creating a new profile")
	}

	if git.BranchExists(repoDir, name) {
		return fmt.Errorf("profile %q already exists", name)
	}

	if !git.BranchExists(repoDir, "default") {
		return fmt.Errorf("'default' branch not found — was this repo initialised with 'claude-profiles init'?")
	}

	action, err := prompt.HandleDirty(repoDir, false)
	if err != nil {
		return err
	}
	if action == prompt.ActionCancel {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := git.CreateBranchFrom(repoDir, name, "default"); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := profile.Restore(repoDir, home); err != nil {
		return err
	}

	fmt.Printf("Created fresh profile %q\n", name)
	return nil
}
