package claude

import (
	"os/exec"
)

// IsRunning returns true if a process named "claude" is currently running.
func IsRunning() (bool, error) {
	err := exec.Command("pgrep", "-x", "claude").Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
