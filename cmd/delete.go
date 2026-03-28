package cmd

import (
	"fmt"

	"github.com/alimoeeny/claude-profiles/internal/config"
	"github.com/alimoeeny/claude-profiles/internal/profile"
	"github.com/alimoeeny/claude-profiles/internal/prompt"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <profile>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	target := args[0]

	storeDir, err := requireStore()
	if err != nil {
		return err
	}

	if target == "default" {
		return fmt.Errorf("cannot delete 'default' — it is the base for create_fresh")
	}

	cfg, err := config.Load(storeDir)
	if err != nil {
		return err
	}
	if cfg.Current == target {
		return fmt.Errorf("cannot delete active profile — switch to another profile first")
	}

	if !profile.ProfileExists(storeDir, target) {
		return fmt.Errorf("profile %q not found", target)
	}

	ok, err := prompt.Confirm(fmt.Sprintf("Delete profile %q? [y/N]: ", target))
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := profile.DeleteProfile(storeDir, target); err != nil {
		return err
	}

	fmt.Printf("Deleted profile %q\n", target)
	return nil
}
