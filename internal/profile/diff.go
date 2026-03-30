package profile

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// DiffStatus classifies how a file has changed relative to a snapshot.
type DiffStatus int

const (
	DiffAdded     DiffStatus = iota // present in live, absent in snapshot
	DiffRemoved                     // present in snapshot, absent in live
	DiffModified                    // present in both, content differs
	DiffUnchanged                   // identical in both
)

// DiffEntry describes one file's change status.
type DiffEntry struct {
	Path     string     // normalized relative path, e.g. ".claude/settings.json"
	Status   DiffStatus
	LiveSize int64 // 0 for Removed
	ZipSize  int64 // 0 for Added
}

// DiffResult holds the full comparison between live home and a snapshot.
type DiffResult struct {
	Entries                          []DiffEntry
	Added, Removed, Modified, Unchanged int
}

// DiffSnapshot compares the live home state against the named profile's zip snapshot.
// Returns an error if the snapshot zip does not exist.
func DiffSnapshot(homeDir, storeDir, name string) (*DiffResult, error) {
	zp := zipPath(storeDir, name)
	if _, err := os.Stat(zp); os.IsNotExist(err) {
		return nil, fmt.Errorf("snapshot for profile %q not found — run 'claude-profiles list' to diagnose", name)
	}

	tmpDir, err := os.MkdirTemp("", "claude-profiles-diff-*")
	if err != nil {
		return nil, fmt.Errorf("diff: create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := extractZipToDir(zp, tmpDir); err != nil {
		return nil, fmt.Errorf("diff: extract snapshot: %w", err)
	}

	liveHashes, liveSizes, err := hashHomeTree(homeDir)
	if err != nil {
		return nil, fmt.Errorf("diff: hash live home: %w", err)
	}

	snapHashes, snapSizes, err := hashHomeTree(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("diff: hash snapshot: %w", err)
	}

	return buildDiffResult(liveHashes, liveSizes, snapHashes, snapSizes), nil
}

// extractZipToDir extracts all files from zipPath into destDir, preserving
// relative paths. Directory entries in the zip are skipped.
func extractZipToDir(zp, destDir string) error {
	r, err := zip.OpenReader(zp)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("extractZipToDir: open %s: %w", zp, err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if err := extractFile(f, filepath.Join(destDir, filepath.FromSlash(f.Name))); err != nil {
			return fmt.Errorf("extractZipToDir %s: %w", f.Name, err)
		}
	}
	return nil
}

// hashHomeTree builds a per-file SHA-256 map and size map for ~/.claude.json
// and all files under ~/.claude/, using the same normalized path keys used in
// zip archives (forward slashes, e.g. ".claude/settings.json").
func hashHomeTree(homeDir string) (hashes map[string][32]byte, sizes map[string]int64, err error) {
	hashes = make(map[string][32]byte)
	sizes = make(map[string]int64)

	jsonPath := filepath.Join(homeDir, ".claude.json")
	if data, e := os.ReadFile(jsonPath); e == nil {
		hashes[".claude.json"] = sha256.Sum256(data)
		sizes[".claude.json"] = int64(len(data))
	} else if !os.IsNotExist(e) {
		return nil, nil, e
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	if _, e := os.Stat(claudeDir); e == nil {
		walkErr := filepath.WalkDir(claudeDir, func(p string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.Type()&fs.ModeSymlink != 0 {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(claudeDir, p)
			if err != nil {
				return err
			}
			key := ".claude/" + filepath.ToSlash(rel)
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			hashes[key] = sha256.Sum256(data)
			sizes[key] = int64(len(data))
			return nil
		})
		if walkErr != nil {
			return nil, nil, walkErr
		}
	}

	return hashes, sizes, nil
}

// buildDiffResult compares two hash maps and produces a DiffResult.
func buildDiffResult(
	liveHashes map[string][32]byte, liveSizes map[string]int64,
	snapHashes map[string][32]byte, snapSizes map[string]int64,
) *DiffResult {
	// Collect all unique paths.
	seen := make(map[string]struct{})
	for k := range liveHashes {
		seen[k] = struct{}{}
	}
	for k := range snapHashes {
		seen[k] = struct{}{}
	}

	paths := make([]string, 0, len(seen))
	for k := range seen {
		paths = append(paths, k)
	}
	sort.Strings(paths)

	result := &DiffResult{}
	for _, p := range paths {
		lh, inLive := liveHashes[p]
		sh, inSnap := snapHashes[p]

		var entry DiffEntry
		entry.Path = p

		switch {
		case inLive && !inSnap:
			entry.Status = DiffAdded
			entry.LiveSize = liveSizes[p]
			result.Added++
		case !inLive && inSnap:
			entry.Status = DiffRemoved
			entry.ZipSize = snapSizes[p]
			result.Removed++
		case lh == sh:
			entry.Status = DiffUnchanged
			entry.LiveSize = liveSizes[p]
			entry.ZipSize = snapSizes[p]
			result.Unchanged++
		default:
			entry.Status = DiffModified
			entry.LiveSize = liveSizes[p]
			entry.ZipSize = snapSizes[p]
			result.Modified++
		}

		result.Entries = append(result.Entries, entry)
	}

	return result
}
