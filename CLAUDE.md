# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build ./...                    # build all packages
go test ./...                     # run all tests
go test ./internal/git/... -v     # run a single package's tests
go run main.go <command>          # run without installing
go install .                      # install as claude-profiles binary
```

## Architecture

Thin cobra commands in `cmd/` delegate to internal packages. No business logic lives in `cmd/`.

| Package | Responsibility |
|---|---|
| `internal/git` | All git operations via `exec.Command("git", ...)` |
| `internal/profile` | `Snapshot` (home→repo) and `Restore` (repo→home) file operations |
| `internal/claude` | Process detection via `pgrep -x claude` |
| `internal/config` | Read/write `config.toml` (TOML via BurntSushi/toml) |
| `internal/prompt` | Shared interactive prompts: `HandleDirty`, `Ask`, `Confirm` |
| `internal/repopath` | Resolves repo path from `$CLAUDE_PROFILES_DIR` or `~/.claude-profiles/` |
| `assets` | Embeds default minimal profile via `go:embed all:default` |

## Core Invariant

Every command that touches `~/.claude/` must follow this exact sequence:

1. `claude.IsRunning()` — abort if Claude is running
2. `prompt.HandleDirty()` — offer Save / Discard / Cancel (or Carry Over for duplicate)
3. Git operation (`Checkout`, `CreateBranch`, `CreateBranchFrom`)
4. `profile.Restore()` — which does `rm -f ~/.claude.json && rm -rf ~/.claude/` then copies from repo

## Profiles Repo

`~/.claude-profiles/` is a git repo where each branch is a full snapshot of `~/.claude.json` + `~/.claude/`. `HEAD` = active profile (no separate state file). `config.toml` at repo root stores `backup_remote`. The `default` branch holds the baked-in minimal config and must never be deleted — it is the seed for `create_fresh`.

## Key Design Decisions

- **`go:embed all:default`** — the `all:` prefix is required to embed dot files (`.claude.json`, `.claude/`)
- **`exec.Command` over go-git** — simpler, respects user's git config, operations are straightforward
- **`CLAUDE_PROFILES_DIR` env var** — overrides the default `~/.claude-profiles/` path, used in testing to avoid touching real profiles
