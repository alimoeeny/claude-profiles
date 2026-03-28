package claude

import (
	"os/exec"
)

// RunPgrep is the function used to check for a running process by name.
// It can be replaced in tests to avoid calling the real pgrep binary.
var RunPgrep = func(name string) error {
	return exec.Command("pgrep", "-x", name).Run()
}

// IsRunning returns true if a process named "claude" is currently running.
func IsRunning() (bool, error) {
	err := RunPgrep("claude")
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
