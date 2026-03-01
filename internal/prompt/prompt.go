package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ali/claude-profile-switcher/internal/git"
)

// DirtyAction is what the user chose to do with uncommitted changes.
type DirtyAction int

const (
	ActionSave      DirtyAction = iota // commit changes to current branch
	ActionDiscard                      // throw away changes
	ActionCarryOver                    // carry changes to new branch (duplicate only)
	ActionCancel                       // abort the operation
)

// HandleDirty checks for uncommitted changes and prompts the user if dirty.
// allowCarryOver adds a [C]arry over option (used by duplicate).
// Returns ActionSave (clean or saved), ActionDiscard, ActionCarryOver, or ActionCancel.
func HandleDirty(repoDir string, allowCarryOver bool) (DirtyAction, error) {
	dirty, err := git.IsDirty(repoDir)
	if err != nil {
		return ActionCancel, err
	}
	if !dirty {
		return ActionSave, nil
	}

	stat, err := git.DiffStat(repoDir)
	if err != nil {
		return ActionCancel, err
	}
	fmt.Println("Uncommitted changes in current profile:")
	fmt.Println(stat)

	var promptText string
	if allowCarryOver {
		promptText = "  [S]ave to current branch / [C]arry over to new branch / [A]bort: "
	} else {
		promptText = "  [S]ave / [D]iscard / [C]ancel: "
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(promptText)
		line, _ := reader.ReadString('\n')
		choice := strings.ToLower(strings.TrimSpace(line))

		if allowCarryOver {
			switch choice {
			case "s":
				if err := git.CommitAll(repoDir, autoSaveMessage()); err != nil {
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
				if err := git.CommitAll(repoDir, autoSaveMessage()); err != nil {
					return ActionCancel, err
				}
				return ActionSave, nil
			case "d":
				if err := git.Discard(repoDir); err != nil {
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

func autoSaveMessage() string {
	return "auto-save: " + time.Now().Format("2006-01-02 15:04")
}

// Ask prints a prompt and reads a single line of input.
// Returns empty string on EOF (treated as the user accepting the default).
func Ask(promptText string) (string, error) {
	fmt.Print(promptText)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if errors.Is(err, io.EOF) {
		return strings.TrimSpace(line), nil
	}
	return strings.TrimSpace(line), err
}

// Confirm asks a yes/no question. Returns true only on "y" or "Y".
func Confirm(promptText string) (bool, error) {
	answer, err := Ask(promptText)
	if err != nil {
		return false, err
	}
	return strings.ToLower(answer) == "y", nil
}
