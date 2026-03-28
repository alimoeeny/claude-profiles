package cmd

import (
	"fmt"
	"os"

	"github.com/alimoeeny/claude-profiles/internal/claude"
	"github.com/alimoeeny/claude-profiles/internal/config"
	"github.com/alimoeeny/claude-profiles/internal/profile"
	"github.com/alimoeeny/claude-profiles/internal/prompt"
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

	storeDir, err := requireStore()
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

	if !profile.ProfileExists(storeDir, target) {
		return fmt.Errorf("profile %q not found — run 'claude-profiles list' to see available profiles", target)
	}

	cfg, err := config.Load(storeDir)
	if err != nil {
		return err
	}
	if cfg.Current == target {
		fmt.Printf("Already on profile %q\n", target)
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	action, err := prompt.HandleDirty(home, storeDir, cfg.Current, false)
	if err != nil {
		return err
	}
	if action == prompt.ActionCancel {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := profile.Restore(storeDir, home, target); err != nil {
		return err
	}

	hash, err := profile.HashHome(home)
	if err != nil {
		return err
	}
	cfg.Current = target
	cfg.SnapshotHash = hash
	if err := config.Save(storeDir, cfg); err != nil {
		return err
	}

	fmt.Printf("Switched to profile %q\n", target)
	return nil
}
