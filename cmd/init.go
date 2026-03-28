package cmd

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alimoeeny/claude-profile-manager/assets"
	"github.com/alimoeeny/claude-profile-manager/internal/config"
	"github.com/alimoeeny/claude-profile-manager/internal/profile"
	"github.com/alimoeeny/claude-profile-manager/internal/prompt"
	"github.com/alimoeeny/claude-profile-manager/internal/repopath"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up the profiles store for the first time",
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

	storeDir, err := prompt.Ask(fmt.Sprintf("Profiles store path [%s]: ", defaultPath))
	if err != nil {
		return err
	}
	if storeDir == "" {
		storeDir = defaultPath
	}

	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return fmt.Errorf("create store dir: %w", err)
	}

	// Write the embedded default profile as profiles/default.zip
	if err := writeDefaultZip(storeDir); err != nil {
		return err
	}

	cfg := &config.Config{}

	// Optionally configure backup remote
	remote, err := prompt.Ask("Backup remote path for rsync (leave blank to skip): ")
	if err != nil {
		return err
	}
	if remote != "" {
		cfg.BackupRemote = remote
	}

	// Snapshot current ~/.claude state into a named profile
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

	if profile.ProfileExists(storeDir, profileName) {
		return fmt.Errorf("profile %q already exists", profileName)
	}

	if err := profile.Snapshot(home, storeDir, profileName); err != nil {
		return err
	}

	hash, err := profile.HashHome(home)
	if err != nil {
		return err
	}
	cfg.Current = profileName
	cfg.SnapshotHash = hash

	if err := config.Save(storeDir, cfg); err != nil {
		return err
	}

	fmt.Printf("\nInitialized profiles store at %s\n", storeDir)
	fmt.Printf("Active profile: %s\n", profileName)
	return nil
}

// writeDefaultZip packs the embedded default assets into profiles/default.zip.
func writeDefaultZip(storeDir string) error {
	profilesDir := filepath.Join(storeDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(profilesDir, "default.zip"))
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	return fs.WalkDir(assets.DefaultProfile, "default", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel("default", filePath)
		if err != nil {
			return err
		}
		if rel == "." || d.IsDir() {
			return nil
		}
		data, err := assets.DefaultProfile.ReadFile(filePath)
		if err != nil {
			return err
		}
		fw, err := w.Create(rel)
		if err != nil {
			return err
		}
		_, err = fw.Write(data)
		return err
	})
}
