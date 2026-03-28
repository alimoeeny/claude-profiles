package cmd

import (
	"fmt"
	"os"

	"github.com/alimoeeny/claude-profile-manager/internal/config"
	"github.com/alimoeeny/claude-profile-manager/internal/profile"
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
	storeDir, err := requireStore()
	if err != nil {
		return err
	}

	cfg, err := config.Load(storeDir)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dirty, err := profile.IsDirty(home, cfg.SnapshotHash)
	if err != nil {
		return err
	}

	state := "clean"
	if dirty {
		state = "dirty (unsaved changes)"
	}

	remote := cfg.BackupRemote
	if remote == "" {
		remote = "(not configured)"
	}

	fmt.Printf("Profile: %s\nState:   %s\nRemote:  %s\n", cfg.Current, state, remote)
	return nil
}
