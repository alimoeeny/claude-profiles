package cmd

import (
	"fmt"
	"os"

	"github.com/ali/claude-profile-switcher/internal/claude"
	"github.com/ali/claude-profile-switcher/internal/config"
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

	storeDir, err := requireStore()
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

	if profile.ProfileExists(storeDir, name) {
		return fmt.Errorf("profile %q already exists", name)
	}

	cfg, err := config.Load(storeDir)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	action, err := prompt.HandleDirty(home, storeDir, cfg.Current, true)
	if err != nil {
		return err
	}
	if action == prompt.ActionCancel {
		fmt.Println("Cancelled.")
		return nil
	}

	switch action {
	case prompt.ActionCarryOver:
		// Snapshot current home state directly as the new profile
		if err := profile.Snapshot(home, storeDir, name); err != nil {
			return err
		}
	default:
		// Copy the saved zip (Save already updated it; Discard uses last clean state)
		if err := profile.DuplicateProfile(storeDir, cfg.Current, name); err != nil {
			return err
		}
	}

	hash, err := profile.HashHome(home)
	if err != nil {
		return err
	}
	cfg.Current = name
	cfg.SnapshotHash = hash
	if err := config.Save(storeDir, cfg); err != nil {
		return err
	}

	fmt.Printf("Created and switched to profile %q\n", name)
	return nil
}
