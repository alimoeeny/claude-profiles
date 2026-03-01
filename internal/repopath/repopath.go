package repopath

import (
	"os"
	"path/filepath"
)

const envVar = "CLAUDE_PROFILES_DIR"

// Resolve returns the profiles repo path: $CLAUDE_PROFILES_DIR or ~/.claude-profiles/.
func Resolve() (string, error) {
	if dir := os.Getenv(envVar); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude-profiles"), nil
}
