package cmd

import (
	"fmt"
	"os"

	"github.com/ali/claude-profile-switcher/internal/config"
	"github.com/ali/claude-profile-switcher/internal/profile"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available profiles",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	storeDir, err := requireStore()
	if err != nil {
		return err
	}

	cfg, err := config.Load(storeDir)
	if err != nil {
		return err
	}

	names, err := profile.ListProfiles(storeDir)
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

	for _, name := range names {
		if name == cfg.Current {
			suffix := ""
			if dirty {
				suffix = " (dirty)"
			}
			fmt.Printf("* %s%s\n", name, suffix)
		} else {
			fmt.Printf("  %s\n", name)
		}
	}
	return nil
}
