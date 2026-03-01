package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ali/claude-profile-switcher/internal/git"
)

// initRepo creates a real git repo in a temp dir with an initial commit.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	f := filepath.Join(dir, "README")
	os.WriteFile(f, []byte("hello"), 0644)
	run("add", "README")
	run("commit", "-m", "init")
	return dir
}

func TestCurrentBranch(t *testing.T) {
	dir := initRepo(t)
	branch, err := git.CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if branch != "main" {
		t.Errorf("CurrentBranch() = %q, want %q", branch, "main")
	}
}

func TestIsDirty(t *testing.T) {
	dir := initRepo(t)

	dirty, err := git.IsDirty(dir)
	if err != nil {
		t.Fatalf("IsDirty() error = %v", err)
	}
	if dirty {
		t.Error("IsDirty() = true on clean repo, want false")
	}

	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("change"), 0644)
	dirty, err = git.IsDirty(dir)
	if err != nil {
		t.Fatalf("IsDirty() error = %v", err)
	}
	if !dirty {
		t.Error("IsDirty() = false after adding file, want true")
	}
}

func TestListBranches(t *testing.T) {
	dir := initRepo(t)
	exec.Command("git", "-C", dir, "checkout", "-b", "work").Run()
	exec.Command("git", "-C", dir, "checkout", "main").Run()

	branches, err := git.ListBranches(dir)
	if err != nil {
		t.Fatalf("ListBranches() error = %v", err)
	}
	if len(branches) != 2 {
		t.Errorf("ListBranches() = %v, want 2 branches", branches)
	}
}

func TestCommitAll(t *testing.T) {
	dir := initRepo(t)
	os.WriteFile(filepath.Join(dir, "change.txt"), []byte("hi"), 0644)

	if err := git.CommitAll(dir, "test commit"); err != nil {
		t.Fatalf("CommitAll() error = %v", err)
	}

	dirty, _ := git.IsDirty(dir)
	if dirty {
		t.Error("repo still dirty after CommitAll()")
	}
}

func TestCheckout(t *testing.T) {
	dir := initRepo(t)
	exec.Command("git", "-C", dir, "checkout", "-b", "work").Run()
	exec.Command("git", "-C", dir, "checkout", "main").Run()

	if err := git.Checkout(dir, "work"); err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}
	branch, _ := git.CurrentBranch(dir)
	if branch != "work" {
		t.Errorf("after Checkout(), branch = %q, want %q", branch, "work")
	}
}

func TestCreateBranch(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateBranch(dir, "new-profile"); err != nil {
		t.Fatalf("CreateBranch() error = %v", err)
	}
	branch, _ := git.CurrentBranch(dir)
	if branch != "new-profile" {
		t.Errorf("after CreateBranch(), branch = %q, want %q", branch, "new-profile")
	}
}

func TestDeleteBranch(t *testing.T) {
	dir := initRepo(t)
	exec.Command("git", "-C", dir, "checkout", "-b", "tobedeleted").Run()
	exec.Command("git", "-C", dir, "checkout", "main").Run()

	if err := git.DeleteBranch(dir, "tobedeleted"); err != nil {
		t.Fatalf("DeleteBranch() error = %v", err)
	}
	branches, _ := git.ListBranches(dir)
	for _, b := range branches {
		if b == "tobedeleted" {
			t.Error("branch still exists after DeleteBranch()")
		}
	}
}

func TestBranchExists(t *testing.T) {
	dir := initRepo(t)
	if !git.BranchExists(dir, "main") {
		t.Error("BranchExists() = false for 'main', want true")
	}
	if git.BranchExists(dir, "nonexistent") {
		t.Error("BranchExists() = true for nonexistent branch, want false")
	}
}

func TestDiffStat(t *testing.T) {
	dir := initRepo(t)
	os.WriteFile(filepath.Join(dir, "change.txt"), []byte("hi"), 0644)

	stat, err := git.DiffStat(dir)
	if err != nil {
		t.Fatalf("DiffStat() error = %v", err)
	}
	if stat == "" {
		t.Error("DiffStat() returned empty string on dirty repo")
	}
}

func TestIsRepo(t *testing.T) {
	dir := initRepo(t)
	if !git.IsRepo(dir) {
		t.Error("IsRepo() = false on valid git repo")
	}
	if git.IsRepo(t.TempDir()) {
		t.Error("IsRepo() = true on non-git directory")
	}
}

func TestDiscard(t *testing.T) {
	dir := initRepo(t)
	os.WriteFile(filepath.Join(dir, "change.txt"), []byte("hi"), 0644)

	if err := git.Discard(dir); err != nil {
		t.Fatalf("Discard() error = %v", err)
	}
	dirty, _ := git.IsDirty(dir)
	if dirty {
		t.Error("repo still dirty after Discard()")
	}
}
