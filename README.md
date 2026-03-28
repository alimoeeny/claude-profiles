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

## How It Works

Each profile is a zip snapshot of `~/.claude.json` and `~/.claude/`. Switching profiles replaces those files entirely — it is not a merge. The tool refuses to switch if Claude is currently running, preventing any risk of corruption.

```
~/.claude-profiles/
├── config.toml          ← tracks active profile + optional backup remote
└── profiles/
    ├── default.zip      ← built-in minimal config (seed for create_fresh)
    ├── main.zip         ← snapshot of ~/.claude.json + ~/.claude/
    └── work.zip

~/.claude/               ← fully replaced on every switch
~/.claude.json           ← fully replaced on every switch
```

> **Note**: Always close Claude Code before switching profiles. The tool will abort with an error if it detects Claude is running.

## Commands Reference

| Command | Arguments | What it does | Notes |
|---|---|---|---|
| `init` | — | Creates the profile store, snapshots current `~/.claude` as your first profile | Run once. |
| `list` | — | Lists all profiles; marks active with `*`, unsaved changes with `(dirty)` | |
| `status` | — | Shows active profile name, clean/dirty state, and backup remote | |
| `switch` | `<name>` | Switches to the named profile | Aborts if Claude is running. Prompts save/discard/cancel for unsaved changes. |
| `create_fresh` | `<name>` | Creates a new blank profile from the built-in default template | Does not switch automatically. |
| `duplicate` | `<name>` | Clones the current profile into a new named profile and switches to it | Prompts to carry over or discard unsaved changes. |
| `delete` | `<name>` | Permanently deletes a profile | Cannot delete the active profile or `default`. |
| `backup` | — | Creates a timestamped `.zip` of all profiles in your home directory | Non-destructive; does not modify the store. |
| `push` | — | Rsyncs the profiles store to the configured remote path | Requires `backup_remote` in `config.toml`. |

## Configuration

The store is configured via `~/.claude-profiles/config.toml`, created automatically by `init`. You can edit it directly at any time.

```toml
# ~/.claude-profiles/config.toml

current       = "work"                              # active profile (managed by the tool)
snapshot_hash = "abc123..."                         # change-detection hash (managed by the tool)
backup_remote = "user@host:/backups/claude-profiles" # rsync destination for `push` (optional)
```

The only field you'd typically set manually is `backup_remote`. You can also configure it interactively during `init` or when running `push` for the first time with no remote set.

## Typical Workflows

### Separate work and personal configs

You use Claude Code at your day job with company-specific `CLAUDE.md` instructions and MCP servers. At home you want a clean slate.

```bash
claude-profiles init                   # saves current config as "main"
claude-profiles duplicate work         # clone current setup as "work"
claude-profiles create_fresh personal  # blank profile for personal use
claude-profiles switch personal        # switch to it
```

### Experimenting with a new MCP server

You want to try a new MCP configuration without risking your working setup.

```bash
claude-profiles duplicate experiment   # clone current profile
# ... make changes inside Claude Code ...
# if it works: keep using it
# if it breaks:
claude-profiles switch work            # discard and return to known-good
```

### Backing up and syncing across machines

```bash
claude-profiles backup                 # save a local timestamped .zip to ~/
claude-profiles push                   # rsync the store to a remote path
```
