package claude

import (
	"os/exec"
	"testing"
)

func TestIsRunning_WhenProcessRunning(t *testing.T) {
	orig := runPgrep
	defer func() { runPgrep = orig }()
	runPgrep = func(name string) error { return nil }

	running, err := IsRunning()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !running {
		t.Fatal("expected IsRunning() = true, got false")
	}
}

func TestIsRunning_WhenProcessNotRunning(t *testing.T) {
	orig := runPgrep
	defer func() { runPgrep = orig }()
	// pgrep exits 1 when no process matches — simulate with "sh -c exit 1"
	runPgrep = func(name string) error {
		return exec.Command("sh", "-c", "exit 1").Run()
	}

	running, err := IsRunning()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if running {
		t.Fatal("expected IsRunning() = false, got true")
	}
}

func TestIsRunning_OnUnexpectedError(t *testing.T) {
	orig := runPgrep
	defer func() { runPgrep = orig }()
	// pgrep exits 2 on usage error — IsRunning should propagate this as an error
	runPgrep = func(name string) error {
		return exec.Command("sh", "-c", "exit 2").Run()
	}

	running, err := IsRunning()
	if err == nil {
		t.Fatal("expected non-nil error for unexpected pgrep exit code")
	}
	if running {
		t.Fatal("expected IsRunning() = false on error")
	}
}
