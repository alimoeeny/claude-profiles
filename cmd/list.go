package cmd

import (
	"fmt"

	"github.com/ali/claude-profile-switcher/internal/git"
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
	repoDir, err := requireRepo()
	if err != nil {
		return err
	}

	current, err := git.CurrentBranch(repoDir)
	if err != nil {
		return err
	}

	branches, err := git.ListBranches(repoDir)
	if err != nil {
		return err
	}

	dirty, err := git.IsDirty(repoDir)
	if err != nil {
		return err
	}

	for _, b := range branches {
		if b == current {
			suffix := ""
			if dirty {
				suffix = " (dirty)"
			}
			fmt.Printf("* %s%s\n", b, suffix)
		} else {
			fmt.Printf("  %s\n", b)
		}
	}
	return nil
}
