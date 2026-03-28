package claude_test

import (
	"testing"

	"github.com/alimoeeny/claude-profiles/internal/claude"
)

func TestIsRunning_ReturnsBool(t *testing.T) {
	// We can't control whether claude is actually running in CI,
	// but we can verify the function returns without error.
	running, err := claude.IsRunning()
	if err != nil {
		t.Fatalf("IsRunning() error = %v", err)
	}
	_ = running
}
