# claude-profiles

> Instantly switch between named Claude Code configurations — work, personal, per-project — without losing your settings.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
![Platform](https://img.shields.io/badge/platform-macOS-lightgrey)
![License](https://img.shields.io/badge/license-MIT-green)

If you use Claude Code across multiple contexts — a day job, side projects, open-source work — you've probably wished you could keep separate `CLAUDE.md` files, different MCP server setups, or distinct settings without manually copying files around. `claude-profiles` treats each configuration as a named snapshot you can switch with one command.

## Demo

```
$ claude-profiles list
  default
  personal
* work         (dirty)

$ claude-profiles switch personal
Active profile has unsaved changes.
  [s] Save changes to "work"
  [d] Discard changes
  [c] Cancel
Choice: s
Saved. Switched to "personal".

$ claude-profiles list
  default
* personal
  work
```

## Installation

```bash
go install github.com/alimoeeny/claude-profile-manager@latest
```

Verify it worked:

```bash
claude-profiles --help
```

> **Note**: Requires Go 1.21 or later. Install Go at [go.dev/dl](https://go.dev/dl).

If `claude-profiles` is not found after installing, make sure your Go bin directory is on your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add that line to your `~/.zshrc` or `~/.bashrc` to make it permanent.

## Quick Start

**1. Run `init` once** to create the profiles store and save your current `~/.claude` config as your first profile:

```bash
claude-profiles init
# Profiles store path [~/.claude-profiles]: ↵
# Name for your current profile [main]: ↵
# Initialized profiles store at ~/.claude-profiles/
# Active profile: main
```

**2. Create a second profile** — either a blank one or a clone of your current setup:

```bash
claude-profiles create_fresh work   # blank slate from the default template
# or
claude-profiles duplicate work      # exact copy of your current profile
```

**3. Switch to it:**

```bash
claude-profiles switch work
```

**4. When you're done, switch back** — the tool will ask what to do with any unsaved changes:

```bash
claude-profiles switch main
# Active profile has unsaved changes.
#   [s] Save changes to "work"
#   [d] Discard changes
#   [c] Cancel
# Choice: s
# Saved. Switched to "main".
```

---
