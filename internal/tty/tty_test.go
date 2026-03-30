package tty

import (
	"os"
	"strings"
	"testing"
)

func TestIsTTY_Pipe(t *testing.T) {
	r, _, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if IsTTY(r) {
		t.Error("expected IsTTY to return false for a pipe")
	}
}

func TestIsTTY_NoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// Even if somehow called with a real TTY fd, NO_COLOR must win.
	if IsTTY(os.Stderr) {
		t.Error("expected IsTTY to return false when NO_COLOR is set")
	}
}

func TestIsTTY_TermDumb(t *testing.T) {
	t.Setenv("TERM", "dumb")
	if IsTTY(os.Stderr) {
		t.Error("expected IsTTY to return false when TERM=dumb")
	}
}

func TestClaudeRunningMessage_NonTTY(t *testing.T) {
	r, _, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	msg := ClaudeRunningMessage(r)

	if strings.Contains(msg, "\033[") {
		t.Error("expected no ANSI escape codes in non-TTY output")
	}
	if !strings.Contains(msg, "[!]") {
		t.Error("expected [!] icon in message")
	}
	if !strings.Contains(msg, "Claude is running") {
		t.Error("expected 'Claude is running' in message")
	}
	if !strings.Contains(msg, "Tip:") {
		t.Error("expected hint line starting with 'Tip:' in message")
	}
	if !strings.Contains(msg, "quit") {
		t.Error("expected actionable word 'quit' in message")
	}
}
