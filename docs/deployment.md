# Deployment Guide

## How Releases Work

Releases are triggered by pushing a git tag. Everything else — building binaries, creating the GitHub Release, and updating the Homebrew formula — is fully automated.

```bash
git tag v1.2.3 && git push origin v1.2.3
```

That single command kicks off the following sequence:

1. GitHub Actions builds `claude-profiles` for `darwin/amd64` (Intel) and `darwin/arm64` (Apple Silicon)
2. Tarballs and a `checksums.txt` are uploaded to a new GitHub Release
3. The `alimoeeny/homebrew-tap` formula is automatically updated with the new version and SHA256 checksums

Users can then upgrade with:

```bash
brew upgrade claude-profiles
```

## Rules

- **Always tag from `main`**. The release workflow enforces this — it will abort if the tagged commit is not on `main`.
- **Use semantic versioning**: `vMAJOR.MINOR.PATCH` (e.g. `v1.0.0`, `v1.2.3`)
  - `PATCH` — bug fixes, no behaviour changes
  - `MINOR` — new commands or features, backwards compatible
  - `MAJOR` — breaking changes
- **Tags are permanent**. Never delete or move a published tag — users may have that version installed.

## Pre-releases

Tags with a pre-release suffix (e.g. `v1.0.0-beta.1`) are automatically marked as pre-release on GitHub and are **not** picked up by `brew install` by default. Use these for testing before a stable release.

## Checklist Before Tagging

- [ ] All changes merged to `main`
- [ ] `go test ./...` passes locally
- [ ] `CHANGELOG` or release notes drafted (GoReleaser generates one from commit messages)
- [ ] Version number decided and not already used (`git tag` to check existing tags)

## What GoReleaser Does

The `.goreleaser.yaml` at the repo root controls the build. It:
- Runs `go mod tidy` and `go test ./...` before building
- Compiles with `CGO_ENABLED=0` for fully static binaries
- Injects the tag as the version string (visible via `claude-profiles --version`)
- Excludes `docs:`, `test:`, and `chore:` commits from the changelog

## Local Testing

You can test the release pipeline locally without pushing a tag using GoReleaser's snapshot mode (requires `goreleaser` installed):

```bash
goreleaser release --snapshot --clean
```

Artifacts land in `dist/` — inspect the tarballs to verify the binary name and contents before cutting a real release.
