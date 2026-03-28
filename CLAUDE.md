# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build ./...                      # build all packages
go test ./...                       # run all tests
go test ./internal/profile/... -v   # run a single package's tests
go run main.go <command>            # run without installing
go install .                        # install as claude-profiles binary
```

## Architecture

Thin cobra commands in `cmd/` delegate to internal packages. No business logic lives in `cmd/`.

| Package | Responsibility |
|---|---|
| `internal/profile` | `Snapshot` (home→zip), `Restore` (zip→home), `HashHome`, `IsDirty`, `ListProfiles`, `DeleteProfile`, `DuplicateProfile` |
| `internal/claude` | Process detection via `pgrep -x claude` |
| `internal/config` | Read/write `config.toml` (TOML via BurntSushi/toml) |
| `internal/prompt` | Shared interactive prompts: `HandleDirty`, `Ask`, `Confirm` |
| `internal/repopath` | Resolves store path from `$CLAUDE_PROFILES_DIR` or `~/.claude-profiles/` |
| `assets` | Embeds default minimal profile via `go:embed all:default` |

## Core Invariant

Every command that touches `~/.claude/` must follow this exact sequence:

1. `claude.IsRunning()` — abort if Claude is running
2. `prompt.HandleDirty()` — offer Save / Discard / Cancel (or Carry Over for duplicate)
3. `profile.Restore()` — which does `rm -f ~/.claude.json && rm -rf ~/.claude/` then extracts the target zip

## Profiles Store

`~/.claude-profiles/` is a plain directory (not a git repo). Each profile is a zip file under `profiles/<name>.zip` containing a snapshot of `~/.claude.json` + `~/.claude/`. The active profile and dirty-state hash are tracked in `config.toml` (no git HEAD). The `default.zip` holds the baked-in minimal config and must never be deleted — it is the seed for `create_fresh`.

```
~/.claude-profiles/
├── config.toml          ← current, snapshot_hash, backup_remote
└── profiles/
    ├── default.zip      ← built-in minimal config (go:embed seed)
    ├── main.zip
    └── work.zip
```

## Key Design Decisions

- **Zip archives over git branches** — each profile is a self-contained `.zip`; no git dependency, no branch management overhead
- **SHA-256 hash for dirty detection** — `config.toml` stores `snapshot_hash`; `profile.HashHome` recomputes on demand; no git index needed
- **`go:embed all:default`** — the `all:` prefix is required to embed dot files (`.claude.json`, `.claude/`)
- **`CLAUDE_PROFILES_DIR` env var** — overrides the default `~/.claude-profiles/` path, used in testing to avoid touching real profiles
