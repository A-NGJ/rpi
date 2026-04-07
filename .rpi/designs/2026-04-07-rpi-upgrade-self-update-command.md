---
date: 2026-04-07T18:47:30+02:00
status: complete
tags:
    - design
topic: rpi upgrade self-update command
---

# Design: rpi upgrade self-update command

## Summary

Add an `rpi upgrade` command that checks GitHub Releases for a newer version and, if found, downloads, verifies, and replaces the current binary in place. Users run it manually whenever they want the latest version — no automatic notifications.

## Context

Today the only way to update the `rpi` binary is to re-run `curl | bash` from install.sh or `go install @latest`. Neither is discoverable from the CLI itself. Users have no way to know a new version exists or update with a single command.

The install script (`install.sh`) already solves platform detection, download, checksum verification, and installation — but from the shell. `rpi upgrade` replicates this logic in Go so the binary is fully self-contained.

## Constraints

- Must work on macOS and Linux (darwin/linux × amd64/arm64) — same platforms as goreleaser targets
- Must verify SHA256 checksums before replacing the binary
- Must not require sudo by default (follow install.sh fallback: `/usr/local/bin` → `~/.local/bin`)
- Must handle dev builds gracefully (version="dev" means unknown version — always offer to upgrade)
- Must not break `go install` users — the command simply replaces whatever binary is running

## Components

### GitHub Release Client

A small `internal/upgrade` package that:
1. Fetches `https://api.github.com/repos/A-NGJ/rpi/releases/latest` (unauthenticated)
2. Parses the JSON response to extract `tag_name` and asset URLs
3. Only uses stdlib (`net/http`, `encoding/json`) — no new dependencies

### Version Comparison

Compare the embedded `version` variable (`cmd/rpi/main.go:14`) against the remote `tag_name`. Approach:
- Strip `v` prefix from both, split on `.`, compare major/minor/patch numerically
- If current version is `dev`, always treat remote as newer
- No third-party semver library needed — the format is strictly `vX.Y.Z`

### Download and Verify

Replicates install.sh logic in Go:
1. Determine OS (`runtime.GOOS`) and arch (`runtime.GOARCH`) — already normalized to goreleaser's naming
2. Download `rpi_{version}_{os}_{arch}.tar.gz` from the release assets
3. Download `checksums.txt` from the same release
4. Compute SHA256 of the archive, compare against expected checksum
5. Extract the `rpi` binary from the tarball

### Binary Replacement

1. Locate the currently running binary via `os.Executable()` (resolves symlinks)
2. Write the new binary to a temp file in the same directory (ensures same filesystem for atomic rename)
3. `os.Rename` the temp file over the current binary (atomic on POSIX)
4. Set executable permissions (`0755`)

This avoids the "binary replacing itself while running" issue — on Unix, the running process keeps its file descriptor; the inode is unlinked only after the process exits.

### Cobra Command

New file `cmd/rpi/upgrade_cmd.go`:
- `Use: "upgrade"`
- `Short: "Upgrade rpi binary to the latest release"`
- Prints current version, fetches latest, compares, downloads if newer, replaces
- If already at latest: prints "Already up to date" and exits 0

## File Structure

```
internal/upgrade/
  upgrade.go        # GitHub API client, download, checksum, replace logic
cmd/rpi/
  upgrade_cmd.go    # Cobra command wiring
```

## Risks

| Risk | Mitigation |
|------|------------|
| GitHub API rate limit (60/hr unauthenticated) | Single call per `rpi upgrade` invocation — unlikely to hit |
| Download fails mid-transfer | Write to temp file first; original binary untouched until verified |
| Checksum mismatch (corrupted download or MITM) | Abort with error, original binary preserved |
| No write permission to binary location | Detect and suggest running with appropriate permissions |
| Binary installed via `go install` to GOBIN | `os.Executable()` resolves to correct path regardless of install method |

## Out of Scope

- Automatic version check notifications on other commands
- Caching of version checks
- Downgrading to older versions
- Upgrading to a specific version (always latest)
- Windows support (not supported by install.sh either)
- Homebrew tap or other package manager integration

## References

- `install.sh` — existing shell-based install/update logic
- `.rpi/specs/distribution.md` — current distribution spec (to be updated)
- `cmd/rpi/main.go:13-16` — embedded version variables
- `.goreleaser.yml` — release artifact naming and build config
