package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/alimoeeny/claude-profiles/internal/config"
	"github.com/alimoeeny/claude-profiles/internal/profile"
)

// DirtyAction is what the user chose to do with unsaved changes.
type DirtyAction int

const (
	ActionSave      DirtyAction = iota // snapshot current home to current profile
	ActionDiscard                      // restore home from current profile zip
	ActionCarryOver                    // carry current home state to new profile (duplicate only)
	ActionCancel                       // abort the operation
)

// HandleDirty checks for unsaved changes and prompts the user if dirty.
// allowCarryOver adds a [C]arry over option (used by duplicate).
// Returns ActionSave (clean or saved), ActionDiscard, ActionCarryOver, or ActionCancel.
func HandleDirty(r io.Reader, homeDir, storeDir, currentProfile string, allowCarryOver bool) (DirtyAction, error) {
	cfg, err := config.Load(storeDir)
	if err != nil {
		return ActionCancel, err
	}

	dirty, err := profile.IsDirty(homeDir, cfg.SnapshotHash)
	if err != nil {
		return ActionCancel, err
	}
	if !dirty {
		return ActionSave, nil
	}

	fmt.Println("Current profile has unsaved changes.")

	var promptText string
	if allowCarryOver {
		promptText = "  [S]ave to current profile / [C]arry over to new profile / [A]bort: "
	} else {
		promptText = "  [S]ave / [D]iscard / [C]ancel: "
	}

	reader := bufio.NewReader(r)
	for {
		fmt.Print(promptText)
		line, _ := reader.ReadString('\n')
		choice := strings.ToLower(strings.TrimSpace(line))

		if allowCarryOver {
			switch choice {
			case "s":
				if err := profile.Snapshot(homeDir, storeDir, currentProfile); err != nil {
					return ActionCancel, err
				}
				if err := UpdateSnapshotHash(homeDir, storeDir); err != nil {
					return ActionCancel, err
				}
				return ActionSave, nil
			case "c":
				return ActionCarryOver, nil
			case "a":
				return ActionCancel, nil
			}
		} else {
			switch choice {
			case "s":
				if err := profile.Snapshot(homeDir, storeDir, currentProfile); err != nil {
					return ActionCancel, err
				}
				if err := UpdateSnapshotHash(homeDir, storeDir); err != nil {
					return ActionCancel, err
				}
				return ActionSave, nil
			case "d":
				if err := profile.Restore(storeDir, homeDir, currentProfile); err != nil {
					return ActionCancel, err
				}
				if err := UpdateSnapshotHash(homeDir, storeDir); err != nil {
					return ActionCancel, err
				}
				return ActionDiscard, nil
			case "c":
				return ActionCancel, nil
			}
		}
		fmt.Println("  Invalid choice.")
	}
}

// UpdateSnapshotHash recomputes the home hash and persists it to config.
func UpdateSnapshotHash(homeDir, storeDir string) error {
	hash, err := profile.HashHome(homeDir)
	if err != nil {
		return err
	}
	cfg, err := config.Load(storeDir)
	if err != nil {
		return err
	}
	cfg.SnapshotHash = hash
	return config.Save(storeDir, cfg)
}

// Ask prints a prompt and reads a single line of input.
// Returns empty string on EOF (treated as the user accepting the default).
func Ask(r io.Reader, promptText string) (string, error) {
	fmt.Print(promptText)
	reader := bufio.NewReader(r)
	line, err := reader.ReadString('\n')
	if errors.Is(err, io.EOF) {
		return strings.TrimSpace(line), nil
	}
	return strings.TrimSpace(line), err
}

// Confirm asks a yes/no question. Returns true only on "y" or "Y".
func Confirm(r io.Reader, promptText string) (bool, error) {
	answer, err := Ask(r, promptText)
	if err != nil {
		return false, err
	}
	return strings.ToLower(answer) == "y", nil
}
