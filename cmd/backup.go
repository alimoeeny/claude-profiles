package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a timestamped zip snapshot of all profiles",
	RunE:  runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func runBackup(cmd *cobra.Command, args []string) error {
	storeDir, err := requireStore()
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02-1504")
	zipPath := filepath.Join(home, fmt.Sprintf(".claude-profiles-backup-%s.zip", timestamp))

	if err := zipDir(storeDir, zipPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Printf("Backup saved to %s\n", zipPath)
	return nil
}

func zipDir(srcDir, destZip string) error {
	f, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	return filepath.WalkDir(srcDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, filePath)
		if err != nil {
			return err
		}
		fw, err := w.Create(rel)
		if err != nil {
			return err
		}
		in, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer in.Close()
		_, err = io.Copy(fw, in)
		return err
	})
}
