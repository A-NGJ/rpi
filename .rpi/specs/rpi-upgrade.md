---
domain: distribution
feature: rpi-upgrade
last_updated: 2026-04-07T18:47:30+02:00
updated_by: .rpi/designs/2026-04-07-rpi-upgrade-self-update-command.md
---

# rpi upgrade

## Purpose

Allow users to update the `rpi` binary to the latest release with a single command. The command checks GitHub Releases, downloads the correct binary for the user's platform, verifies its checksum, and replaces the current binary in place.

## Scenarios

### Upgrading to a newer version
Given the installed binary is at version v1.0.0
And a newer version v1.1.0 is available on GitHub Releases
When the user runs `rpi upgrade`
Then the new binary is downloaded, checksum-verified, and replaces the current one
And the output confirms the upgrade from v1.0.0 to v1.1.0

### Already at the latest version
Given the installed binary version matches the latest GitHub Release
When the user runs `rpi upgrade`
Then the output says the binary is already up to date
And the binary is not modified

### Upgrading from a development build
Given the binary was built from source without release tags (version is "dev")
When the user runs `rpi upgrade`
Then the latest release is downloaded and installed
And the output confirms the installed version

### Checksum verification failure
Given a download produces a corrupted or tampered archive
When `rpi upgrade` verifies the checksum
Then the upgrade is aborted with a checksum mismatch error
And the current binary is not modified

### No network connectivity
Given the GitHub API or release download is unreachable
When the user runs `rpi upgrade`
Then the command exits with a clear network error
And the current binary is not modified

### Insufficient write permissions
Given the user does not have write permission to the binary's directory
When `rpi upgrade` attempts to replace the binary
Then the command exits with a permissions error suggesting how to fix it

## Constraints
- Only replaces the binary found via the running process's executable path
- SHA256 checksum verification is mandatory — never skip it
- The original binary must remain intact if any step fails (download, verify, extract, replace)
- No new third-party dependencies — use Go stdlib only
- Supports macOS and Linux (darwin/linux × amd64/arm64)

## Out of Scope
- Automatic upgrade checks on other commands
- Downgrading or pinning to a specific version
- Windows support
- Self-update for `go install` users (they use `go install @latest`)
