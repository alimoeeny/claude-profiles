package cmd

import (
	"fmt"

	"github.com/ali/claude-profile-switcher/internal/git"
	"github.com/ali/claude-profile-switcher/internal/prompt"
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

	repoDir, err := requireRepo()
	if err != nil {
		return err
	}

	if target == "default" {
		return fmt.Errorf("cannot delete 'default' — it is the base for create_fresh")
	}

	current, err := git.CurrentBranch(repoDir)
	if err != nil {
		return err
	}
	if current == target {
		return fmt.Errorf("cannot delete active profile — switch to another profile first")
	}

	if !git.BranchExists(repoDir, target) {
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

	if err := git.DeleteBranch(repoDir, target); err != nil {
		return err
	}

	fmt.Printf("Deleted profile %q\n", target)
	return nil
}
