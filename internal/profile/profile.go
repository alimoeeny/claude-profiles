package profile

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const profilesSubdir = "profiles"

func profilesDir(storeDir string) string {
	return filepath.Join(storeDir, profilesSubdir)
}

func zipPath(storeDir, name string) string {
	return filepath.Join(profilesDir(storeDir), name+".zip")
}

// Snapshot zips ~/.claude.json and ~/.claude/ from homeDir into storeDir/profiles/<name>.zip.
func Snapshot(homeDir, storeDir, name string) error {
	if err := os.MkdirAll(profilesDir(storeDir), 0755); err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}

	f, err := os.Create(zipPath(storeDir, name))
	if err != nil {
		return fmt.Errorf("snapshot: create zip: %w", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	if err := addFileToZip(w, filepath.Join(homeDir, ".claude.json"), ".claude.json"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("snapshot .claude.json: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	if _, err := os.Stat(claudeDir); err == nil {
		if err := addDirToZip(w, claudeDir, ".claude"); err != nil {
			return fmt.Errorf("snapshot .claude/: %w", err)
		}
	}

	return nil
}

// Restore extracts storeDir/profiles/<name>.zip into homeDir, replacing ~/.claude.json and ~/.claude/.
func Restore(storeDir, homeDir, name string) error {
	os.Remove(filepath.Join(homeDir, ".claude.json"))
	os.RemoveAll(filepath.Join(homeDir, ".claude"))

	zp := zipPath(storeDir, name)
	r, err := zip.OpenReader(zp)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("restore: open %s: %w", zp, err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if f.Name != ".claude.json" && !strings.HasPrefix(f.Name, ".claude/") {
			return fmt.Errorf("restore: unexpected path in zip: %s", f.Name)
		}
		if err := extractFile(f, filepath.Join(homeDir, f.Name)); err != nil {
			return fmt.Errorf("restore %s: %w", f.Name, err)
		}
	}
	return nil
}

// ListProfiles returns all profile names in storeDir/profiles/, sorted alphabetically.
func ListProfiles(storeDir string) ([]string, error) {
	entries, err := os.ReadDir(profilesDir(storeDir))
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".zip" {
			names = append(names, strings.TrimSuffix(e.Name(), ".zip"))
		}
	}
	sort.Strings(names)
	return names, nil
}

// ProfileExists returns true if storeDir/profiles/<name>.zip exists.
func ProfileExists(storeDir, name string) bool {
	_, err := os.Stat(zipPath(storeDir, name))
	return err == nil
}

// DeleteProfile removes storeDir/profiles/<name>.zip.
func DeleteProfile(storeDir, name string) error {
	if err := os.Remove(zipPath(storeDir, name)); err != nil {
		return fmt.Errorf("delete profile %q: %w", name, err)
	}
	return nil
}

// DuplicateProfile copies storeDir/profiles/<src>.zip to storeDir/profiles/<dst>.zip.
func DuplicateProfile(storeDir, src, dst string) error {
	return copyFile(zipPath(storeDir, src), zipPath(storeDir, dst))
}

// HashHome computes a SHA256 fingerprint of ~/.claude.json and all files under ~/.claude/.
// Returns "" if neither exists.
func HashHome(homeDir string) (string, error) {
	type entry struct {
		path    string
		content []byte
	}

	var entries []entry

	jsonPath := filepath.Join(homeDir, ".claude.json")
	if data, err := os.ReadFile(jsonPath); err == nil {
		entries = append(entries, entry{".claude.json", data})
	} else if !os.IsNotExist(err) {
		return "", err
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	if _, err := os.Stat(claudeDir); err == nil {
		walkErr := filepath.WalkDir(claudeDir, func(filePath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type()&fs.ModeSymlink != 0 {
				return nil
			}
			if d.IsDir() && d.Name() == ".git" {
				return filepath.SkipDir
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(claudeDir, filePath)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			entries = append(entries, entry{".claude/" + rel, data})
			return nil
		})
		if walkErr != nil {
			return "", walkErr
		}
	}

	if len(entries) == 0 {
		return "", nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].path < entries[j].path
	})

	h := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(h, "%s\x00", e.path)
		h.Write(e.content)
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// IsDirty returns true if the current home state differs from storedHash.
func IsDirty(homeDir, storedHash string) (bool, error) {
	currentHash, err := HashHome(homeDir)
	if err != nil {
		return false, err
	}
	return currentHash != storedHash, nil
}

func addFileToZip(w *zip.Writer, filePath, name string) error {
	in, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer in.Close()

	fw, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, in)
	return err
}

func addDirToZip(w *zip.Writer, srcDir, prefix string) error {
	return filepath.WalkDir(srcDir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, filePath)
		if err != nil {
			return err
		}
		return addFileToZip(w, filePath, prefix+"/"+filepath.ToSlash(rel))
	})
}

func extractFile(f *zip.File, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, r)
	return err
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
