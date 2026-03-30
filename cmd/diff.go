package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/alimoeeny/claude-profiles/internal/config"
	"github.com/alimoeeny/claude-profiles/internal/profile"
	"github.com/alimoeeny/claude-profiles/internal/tty"
	"github.com/spf13/cobra"
)

var diffVerbose bool
var diffDepth int

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show what has changed in the active profile since the last snapshot",
	RunE:  runDiff,
}

func init() {
	diffCmd.Flags().BoolVarP(&diffVerbose, "verbose", "v", false, "Show full paths and file sizes (ignores --depth)")
	diffCmd.Flags().IntVarP(&diffDepth, "depth", "d", 2, "Max path depth to display; 0 means unlimited")
	rootCmd.AddCommand(diffCmd)
}

func runDiff(cmd *cobra.Command, args []string) error {
	storeDir, err := requireStore()
	if err != nil {
		return err
	}

	cfg, err := config.Load(storeDir)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	result, err := profile.DiffSnapshot(home, storeDir, cfg.Current)
	if err != nil {
		return err
	}

	printDiff(os.Stdout, result, diffVerbose, diffDepth)
	return nil
}

func printDiff(out *os.File, result *profile.DiffResult, verbose bool, depth int) {
	isTerminal := tty.IsTTY(out)

	if result.Added == 0 && result.Removed == 0 && result.Modified == 0 {
		fmt.Fprintln(out, "Profile is clean — no changes since last snapshot.")
		return
	}

	fmt.Fprintf(out, "Changes since last snapshot: %d added, %d removed, %d modified\n",
		result.Added, result.Removed, result.Modified)

	for _, status := range []profile.DiffStatus{profile.DiffAdded, profile.DiffRemoved, profile.DiffModified} {
		if verbose {
			printEntriesVerbose(out, isTerminal, result.Entries, status)
		} else {
			printEntriesGrouped(out, isTerminal, result.Entries, status, depth)
		}
	}
}

// printEntriesVerbose prints every entry for the given status with file sizes.
func printEntriesVerbose(out *os.File, isTerminal bool, entries []profile.DiffEntry, status profile.DiffStatus) {
	for _, e := range entries {
		if e.Status != status {
			continue
		}
		symbol, color := diffSymbol(e.Status)
		line := fmt.Sprintf("  %s %s", symbol, e.Path)
		switch e.Status {
		case profile.DiffAdded:
			line += fmt.Sprintf(" (%d B)", e.LiveSize)
		case profile.DiffRemoved:
			line += fmt.Sprintf(" (%d B in snapshot)", e.ZipSize)
		case profile.DiffModified:
			line += fmt.Sprintf(" (snapshot: %d B → live: %d B)", e.ZipSize, e.LiveSize)
		}
		printLine(out, isTerminal, color, line)
	}
}

// printEntriesGrouped prints entries for the given status, collapsing paths
// that share the same depth-truncated prefix into a single line with a count.
func printEntriesGrouped(out *os.File, isTerminal bool, entries []profile.DiffEntry, status profile.DiffStatus, depth int) {
	// Maintain insertion order by tracking the first time each key appears.
	type group struct {
		displayPath string
		count       int
	}
	var order []string
	groups := map[string]*group{}

	for _, e := range entries {
		if e.Status != status {
			continue
		}
		key := truncatePath(e.Path, depth)
		if _, seen := groups[key]; !seen {
			order = append(order, key)
			groups[key] = &group{displayPath: key}
		}
		groups[key].count++
	}

	symbol, color := diffSymbol(status)
	for _, key := range order {
		g := groups[key]
		line := fmt.Sprintf("  %s %s", symbol, g.displayPath)
		if g.count > 1 {
			line += fmt.Sprintf(" (%d files)", g.count)
		}
		printLine(out, isTerminal, color, line)
	}
}

// truncatePath limits a forward-slash path to depth components.
// depth=0 returns the path unchanged.
func truncatePath(path string, depth int) string {
	if depth == 0 {
		return path
	}
	parts := strings.Split(path, "/")
	if len(parts) <= depth {
		return path
	}
	return strings.Join(parts[:depth], "/")
}

func printLine(out *os.File, isTerminal bool, color, line string) {
	if isTerminal {
		fmt.Fprintln(out, color+line+"\033[0m")
	} else {
		fmt.Fprintln(out, line)
	}
}

// diffSymbol returns the single-character symbol and ANSI color for a DiffStatus.
func diffSymbol(s profile.DiffStatus) (symbol, color string) {
	switch s {
	case profile.DiffAdded:
		return "+", "\033[32m" // green
	case profile.DiffRemoved:
		return "-", "\033[31m" // red
	case profile.DiffModified:
		return "~", "\033[33m" // yellow
	default:
		return " ", ""
	}
}
