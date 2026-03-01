package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// Init initialises a new git repo with the given initial branch name.
func Init(repoDir, initialBranch string) error {
	if IsRepo(repoDir) {
		return nil
	}
	_, err := run(repoDir, "init", "-b", initialBranch)
	return err
}

// CurrentBranch returns the name of the currently checked-out branch.
func CurrentBranch(repoDir string) (string, error) {
	return run(repoDir, "branch", "--show-current")
}

// IsDirty returns true if the working tree has uncommitted changes.
func IsDirty(repoDir string) (bool, error) {
	out, err := run(repoDir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out != "", nil
}

// ListBranches returns all local branch names.
func ListBranches(repoDir string) ([]string, error) {
	out, err := run(repoDir, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// CommitAll stages all changes and creates a commit with the given message.
func CommitAll(repoDir, message string) error {
	if _, err := run(repoDir, "add", "-A"); err != nil {
		return err
	}
	_, err := run(repoDir, "commit", "-m", message)
	return err
}

// Checkout switches to an existing branch.
func Checkout(repoDir, branch string) error {
	_, err := run(repoDir, "checkout", branch)
	return err
}

// CreateBranch creates a new branch from HEAD and switches to it.
func CreateBranch(repoDir, branch string) error {
	_, err := run(repoDir, "checkout", "-b", branch)
	return err
}

// CreateBranchFrom creates a new branch from a specific base branch and switches to it.
func CreateBranchFrom(repoDir, branch, base string) error {
	if err := Checkout(repoDir, base); err != nil {
		return err
	}
	_, err := run(repoDir, "checkout", "-b", branch)
	return err
}

// DeleteBranch deletes a local branch.
func DeleteBranch(repoDir, branch string) error {
	_, err := run(repoDir, "branch", "-d", branch)
	return err
}

// BranchExists returns true if a local branch with the given name exists.
func BranchExists(repoDir, branch string) bool {
	_, err := run(repoDir, "rev-parse", "--verify", branch)
	return err == nil
}

// DiffStat returns a human-readable summary of uncommitted changes.
func DiffStat(repoDir string) (string, error) {
	return run(repoDir, "status", "--short")
}

// AddRemote adds a git remote.
func AddRemote(repoDir, name, url string) error {
	_, err := run(repoDir, "remote", "add", name, url)
	return err
}

// PushAll pushes all branches to the given remote.
func PushAll(repoDir, remote string) (string, error) {
	return run(repoDir, "push", remote, "--all")
}

// Discard discards all uncommitted changes (tracked files and untracked files).
func Discard(repoDir string) error {
	if _, err := run(repoDir, "checkout", "--", "."); err != nil {
		return err
	}
	_, err := run(repoDir, "clean", "-fd")
	return err
}

// IsRepo returns true if repoDir is a git repository.
func IsRepo(repoDir string) bool {
	_, err := run(repoDir, "rev-parse", "--git-dir")
	return err == nil
}

// SetConfig sets a git config value in the repo.
func SetConfig(repoDir, key, value string) error {
	_, err := run(repoDir, "config", key, value)
	return err
}
