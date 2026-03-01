package profile

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Snapshot copies ~/.claude.json and ~/.claude/ from homeDir into repoDir.
func Snapshot(homeDir, repoDir string) error {
	src := filepath.Join(homeDir, ".claude.json")
	dst := filepath.Join(repoDir, ".claude.json")
	if err := copyFile(src, dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("snapshot .claude.json: %w", err)
	}

	srcDir := filepath.Join(homeDir, ".claude")
	dstDir := filepath.Join(repoDir, ".claude")
	if err := copyDir(srcDir, dstDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("snapshot .claude/: %w", err)
	}
	return nil
}

// Restore copies .claude.json and .claude/ from repoDir into homeDir,
// removing stale files first.
func Restore(repoDir, homeDir string) error {
	os.Remove(filepath.Join(homeDir, ".claude.json"))
	os.RemoveAll(filepath.Join(homeDir, ".claude"))

	src := filepath.Join(repoDir, ".claude.json")
	dst := filepath.Join(homeDir, ".claude.json")
	if err := copyFile(src, dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("restore .claude.json: %w", err)
	}

	srcDir := filepath.Join(repoDir, ".claude")
	dstDir := filepath.Join(homeDir, ".claude")
	if err := copyDir(srcDir, dstDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("restore .claude/: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip symlinks — they may point to directories and can't be read as files
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		// Skip nested .git directories — they would be treated as submodules
		// by the profile repo, preventing git add -A from staging their contents
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(src, filePath)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(filePath, target)
	})
}
