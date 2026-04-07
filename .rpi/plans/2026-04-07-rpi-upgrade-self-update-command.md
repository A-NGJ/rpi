---
date: 2026-04-07T18:49:58+02:00
design: .rpi/designs/2026-04-07-rpi-upgrade-self-update-command.md
spec: .rpi/specs/rpi-upgrade.md
status: complete
tags:
    - plan
topic: rpi upgrade self-update command
---

# rpi upgrade self-update command — Implementation Plan

## Overview

Add `rpi upgrade` command that checks GitHub Releases for a newer version, downloads the correct platform binary, verifies its SHA256 checksum, and atomically replaces the current binary.

**Scope**: 2 new files, 2 new test files, 1 modified spec

## Source Documents
- **Design**: .rpi/designs/2026-04-07-rpi-upgrade-self-update-command.md
- **Spec**: .rpi/specs/rpi-upgrade.md

## Phase 1: Upgrade library (`internal/upgrade`)

### Overview
Implement the core upgrade logic as a testable library: GitHub API client, version comparison, download with checksum verification, and binary replacement. All spec scenarios are covered here via unit tests with `httptest` servers.

### Tasks:

#### 1. Upgrade package
**File**: `internal/upgrade/upgrade.go`
**Changes**:
- `FetchLatestRelease(repo string) (Release, error)` — HTTP GET to GitHub API `/repos/{repo}/releases/latest`, parse JSON for `tag_name` and asset browser_download_urls
- `CompareVersions(current, latest string) (bool, error)` — strip `v` prefix, split on `.`, compare major/minor/patch numerically. `"dev"` always returns true (needs upgrade)
- `DownloadAndVerify(archiveURL, checksumsURL, expectedArchiveName string) ([]byte, error)` — download archive + checksums.txt, compute SHA256, compare, extract binary from tarball
- `ReplaceBinary(newBinary []byte) error` — resolve `os.Executable()`, write to temp file in same dir, `os.Rename` atomically, `chmod 0755`
- `Upgrade(repo, currentVersion string) error` — orchestrates the above: fetch → compare → download → verify → replace. Prints progress to stdout. Returns nil if already up to date (prints message)
- Use `runtime.GOOS` and `runtime.GOARCH` for platform detection
- Archive name format: `rpi_{version}_{os}_{arch}.tar.gz` (version without `v` prefix)

#### 2. Tests
**File**: `internal/upgrade/upgrade_test.go`
**Changes**:
- `TestCompareVersions` — table-driven: newer available, already current, dev build, malformed versions
- `TestFetchLatestRelease` — httptest server returning mock GitHub API JSON
- `TestDownloadAndVerify_Success` — httptest serving a real tarball + matching checksums.txt
- `TestDownloadAndVerify_ChecksumMismatch` — httptest serving tarball + wrong checksum → error
- `TestFetchLatestRelease_NetworkError` — unreachable server URL → clear error
- `TestReplaceBinary` — write a dummy binary to a temp dir, replace it, verify contents changed

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/upgrade/...` passes
- [x] `go vet ./internal/upgrade/...` clean

### Commit:
- [x] Stage: `internal/upgrade/`
- [x] Message: `feat(upgrade): add internal upgrade package with GitHub release client`

---

## Phase 2: Cobra command + spec update

### Overview
Wire the upgrade library into a Cobra command and update the distribution spec to include upgrade scenarios.

### Tasks:

#### 1. Upgrade command
**File**: `cmd/rpi/upgrade_cmd.go`
**Changes**:
- `upgradeCmd` Cobra command: `Use: "upgrade"`, `Short: "Upgrade rpi binary to the latest release"`
- `RunE` calls `upgrade.Upgrade("A-NGJ/rpi", version)` where `version` is the build-time embedded variable
- Register via `rootCmd.AddCommand(upgradeCmd)` in `init()`

#### 2. Command test
**File**: `cmd/rpi/upgrade_cmd_test.go`
**Changes**:
- Test that the upgrade command is registered and has correct Use/Short fields (following `version_test.go` pattern)

#### 3. Update distribution spec
**File**: `.rpi/specs/distribution.md`
**Changes**:
- Add scenario: "Upgrade command updates binary to latest release"
- Add scenario: "Upgrade command reports when already up to date"
- Remove "self-update" from Out of Scope (it's now in scope via the rpi-upgrade spec)

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` passes (full suite)
- [x] `go build ./cmd/rpi` succeeds
- [x] `go vet ./...` clean

### Commit:
- [x] Stage: `cmd/rpi/upgrade_cmd.go`, `cmd/rpi/upgrade_cmd_test.go`, `.rpi/specs/distribution.md`
- [x] Message: `feat(upgrade): add rpi upgrade command and update distribution spec`

---

## References
- Design: .rpi/designs/2026-04-07-rpi-upgrade-self-update-command.md
- Spec: .rpi/specs/rpi-upgrade.md
- Existing patterns: `cmd/rpi/version.go`, `cmd/rpi/version_test.go`, `install.sh`
