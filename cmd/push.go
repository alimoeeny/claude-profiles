package cmd

import (
	"fmt"

	"github.com/ali/claude-profile-switcher/internal/config"
	"github.com/ali/claude-profile-switcher/internal/git"
	"github.com/ali/claude-profile-switcher/internal/prompt"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push all profiles to the configured git remote",
	RunE:  runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	repoDir, err := requireRepo()
	if err != nil {
		return err
	}

	cfg, err := config.Load(repoDir)
	if err != nil {
		return err
	}

	if cfg.BackupRemote == "" {
		url, err := prompt.Ask("No remote configured. Enter remote URL (or press Enter to skip): ")
		if err != nil {
			return err
		}
		if url == "" {
			fmt.Println("To configure a remote manually:")
			fmt.Printf("  git -C %s remote add origin <url>\n", repoDir)
			fmt.Println("Then re-run: claude-profiles push")
			return nil
		}
		cfg.BackupRemote = url
		if err := config.Save(repoDir, cfg); err != nil {
			return err
		}
		if err := git.AddRemote(repoDir, "origin", url); err != nil {
			return err
		}
		// Commit the updated config — ignore error if nothing changed
		_ = git.CommitAll(repoDir, "chore: add backup remote to config")
	}

	action, err := prompt.HandleDirty(repoDir, false)
	if err != nil {
		return err
	}
	if action == prompt.ActionCancel {
		fmt.Println("Cancelled.")
		return nil
	}

	out, err := git.PushAll(repoDir, "origin")
	if err != nil {
		return err
	}
	if out != "" {
		fmt.Println(out)
	}

	branches, _ := git.ListBranches(repoDir)
	fmt.Printf("Pushed %d profiles to %s\n", len(branches), cfg.BackupRemote)
	return nil
}
