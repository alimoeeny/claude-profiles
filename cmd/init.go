package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ali/claude-profile-switcher/assets"
	"github.com/ali/claude-profile-switcher/internal/config"
	"github.com/ali/claude-profile-switcher/internal/git"
	"github.com/ali/claude-profile-switcher/internal/profile"
	"github.com/ali/claude-profile-switcher/internal/prompt"
	"github.com/ali/claude-profile-switcher/internal/repopath"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up the profiles repository for the first time",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	defaultPath, err := repopath.Resolve()
	if err != nil {
		return err
	}

	repoDir, err := prompt.Ask(fmt.Sprintf("Profiles repo path [%s]: ", defaultPath))
	if err != nil {
		return err
	}
	if repoDir == "" {
		repoDir = defaultPath
	}

	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("create repo dir: %w", err)
	}

	if err := git.Init(repoDir, "default"); err != nil {
		return err
	}

	// Need user identity for commits in fresh repos
	if err := git.SetConfig(repoDir, "user.email", "claude-profiles@local"); err != nil {
		return err
	}
	if err := git.SetConfig(repoDir, "user.name", "Claude Profiles"); err != nil {
		return err
	}

	// Write embedded default config to the default branch
	if err := writeEmbeddedDefault(repoDir); err != nil {
		return err
	}
	cfg := &config.Config{}
	if err := config.Save(repoDir, cfg); err != nil {
		return err
	}
	if err := git.CommitAll(repoDir, "init: default profile"); err != nil {
		return err
	}

	// Optionally configure backup remote
	remote, err := prompt.Ask("Backup remote URL (leave blank to skip): ")
	if err != nil {
		return err
	}
	if remote != "" {
		cfg.BackupRemote = remote
		if err := config.Save(repoDir, cfg); err != nil {
			return err
		}
		if err := git.AddRemote(repoDir, "origin", remote); err != nil {
			return err
		}
		if err := git.CommitAll(repoDir, "chore: add backup remote to config"); err != nil {
			return err
		}
	}

	// Snapshot current ~/.claude state into a named profile branch
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	profileName, err := prompt.Ask("Name for your current profile [main]: ")
	if err != nil {
		return err
	}
	if profileName == "" {
		profileName = "main"
	}

	if err := git.CreateBranch(repoDir, profileName); err != nil {
		return err
	}
	if err := profile.Snapshot(home, repoDir); err != nil {
		return err
	}
	if err := config.Save(repoDir, cfg); err != nil {
		return err
	}
	if err := git.CommitAll(repoDir, "init: snapshot current profile as '"+profileName+"'"); err != nil {
		return err
	}

	fmt.Printf("\nInitialized profiles repo at %s\n", repoDir)
	fmt.Printf("Active profile: %s\n", profileName)
	return nil
}

func writeEmbeddedDefault(repoDir string) error {
	return fs.WalkDir(assets.DefaultProfile, "default", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel("default", filePath)
		if err != nil {
			return err
		}
		// Skip the root "default" dir entry itself
		if rel == "." {
			return nil
		}
		dst := filepath.Join(repoDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}
		data, err := assets.DefaultProfile.ReadFile(filePath)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0644)
	})
}
