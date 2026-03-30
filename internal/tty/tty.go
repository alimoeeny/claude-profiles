// Package tty provides terminal detection and ANSI color helpers.
// TTY detection uses syscall.TIOCGETA and is macOS-only, consistent with
// the rest of the codebase (pgrep -x claude is also macOS-specific).
package tty

import (
	"os"
	"syscall"
	"unsafe"
)

// IsTTY reports whether f is connected to a terminal.
// Returns false when NO_COLOR is set, TERM=dumb, or the fd is not a TTY.
func IsTTY(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	var termios syscall.Termios
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		// TIOCGETA is the macOS constant for getting terminal attributes.
		syscall.TIOCGETA,
		uintptr(unsafe.Pointer(&termios)),
	)
	return errno == 0
}

// ANSI escape codes.
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	yellow = "\033[33m"
)

// ClaudeRunningMessage returns a human-friendly error message for when Claude
// is detected as running. Output is ANSI-colored when f is a TTY.
func ClaudeRunningMessage(f *os.File) string {
	const (
		msg  = "[!] Claude is running — quit Claude before switching profiles."
		hint = "Tip: quit Claude, then re-run your command."
	)
	if IsTTY(f) {
		return bold + red + msg + reset + "\n" + yellow + hint + reset
	}
	return msg + "\n" + hint
}
