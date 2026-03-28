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

var createFreshCmd = &cobra.Command{
	Use:   "create_fresh <name>",
	Short: "Create a new profile from the default template",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateFresh,
}

func init() {
	rootCmd.AddCommand(createFreshCmd)
}

func runCreateFresh(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("Claude is running — close it before creating a new profile")
	}

	if profile.ProfileExists(storeDir, name) {
		return fmt.Errorf("profile %q already exists", name)
	}

	if !profile.ProfileExists(storeDir, "default") {
		return fmt.Errorf("'default' profile not found — was this store initialised with 'claude-profiles init'?")
	}

	cfg, err := config.Load(storeDir)
	if err != nil {
		return err
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

	if err := profile.DuplicateProfile(storeDir, "default", name); err != nil {
		return err
	}

	if err := profile.Restore(storeDir, home, name); err != nil {
		return err
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

	fmt.Printf("Created fresh profile %q\n", name)
	return nil
}
