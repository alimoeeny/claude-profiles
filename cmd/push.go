package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/alimoeeny/claude-profiles/internal/config"
	"github.com/alimoeeny/claude-profiles/internal/prompt"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Sync all profiles to the configured remote path via rsync",
	RunE:  runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	storeDir, err := requireStore()
	if err != nil {
		return err
	}

	cfg, err := config.Load(storeDir)
	if err != nil {
		return err
	}

	if cfg.BackupRemote == "" {
		dest, err := prompt.Ask("No remote configured. Enter rsync destination (or press Enter to skip): ")
		if err != nil {
			return err
		}
		if dest == "" {
			fmt.Println("To configure a remote, re-run with a destination path (e.g. user@host:/backups/claude-profiles/).")
			return nil
		}
		cfg.BackupRemote = dest
		if err := config.Save(storeDir, cfg); err != nil {
			return err
		}
	}

	// rsync the profiles/ directory to the remote destination
	profilesDir := filepath.Join(storeDir, "profiles") + string(filepath.Separator)
	out, err := exec.Command("rsync", "-a", profilesDir, cfg.BackupRemote).CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync failed: %w\n%s", err, out)
	}
	if len(out) > 0 {
		fmt.Print(string(out))
	}

	fmt.Printf("Profiles synced to %s\n", cfg.BackupRemote)
	return nil
}
