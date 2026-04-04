---
archived_date: "2026-04-02"
date: 2026-03-23T00:00:00Z
design: .rpi/designs/2026-03-22-distribution-pipeline.md
spec: .rpi/specs/distribution.md
status: archived
tags:
    - plan
    - distribution
    - release
    - goreleaser
topic: Distribution Pipeline
---

# Plan: Distribution Pipeline

## Goal

Set up goreleaser cross-platform builds, a GitHub Actions release workflow, a curl|bash install script, and an `rpi version` command — so users can install `rpi` without a Go toolchain.

## Phase 1: Version Command

**Spec coverage:** DP-1, DP-2, DP-3

### Tasks

- [x] Add version/commit/date build vars to `cmd/rpi/main.go`
- [x] Create `cmd/rpi/version.go` — `rpi version` subcommand that prints `rpi version <ver> (commit <hash>, built <date>)`
- [x] Add unit test in `cmd/rpi/version_test.go` — verify default dev output contains "dev", "none", "unknown"
- [x] Commit: `feat: add rpi version command with build-time injection`

### Success Criteria

- `go test ./cmd/rpi/...` passes
- `go build -o bin/rpi ./cmd/rpi && ./bin/rpi version` outputs line containing `dev`, `none`, `unknown`
- `go build -ldflags "-X main.version=v0.1.0 -X main.commit=abc1234 -X main.date=2026-03-23" -o bin/rpi ./cmd/rpi && ./bin/rpi version` outputs `v0.1.0`, `abc1234`, `2026-03-23`

## Phase 2: Goreleaser Configuration

**Spec coverage:** DP-4, DP-5, DP-6, DP-7

### Tasks

- [x] Create `.goreleaser.yml` at repo root:
  - Binary name: `rpi`
  - Main: `./cmd/rpi`
  - CGO_ENABLED=0
  - Targets: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
  - Ldflags: `-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}`
  - Archive name_template: `rpi_{{ .Version }}_{{ .Os }}_{{ .Arch }}`
  - Format: tar.gz (zip override for windows)
  - Checksum: SHA256
- [x] Commit: `feat: add goreleaser configuration`

### Success Criteria

- `goreleaser check` exits 0 (validates config without errors)
- Manual: review that targets, archive format, and name template match design

## Phase 3: GitHub Actions Release Workflow

**Spec coverage:** DP-8, DP-9

### Tasks

- [x] Create `.github/workflows/release.yml`:
  - Trigger: `push: tags: ["v*"]`
  - Jobs: checkout (fetch-depth 0) → setup Go → run goreleaser-action
  - Permissions: `contents: write` (for creating releases)
  - Uses `GITHUB_TOKEN` (auto-provided)
- [x] Commit: `ci: add release workflow for goreleaser`

### Success Criteria

- YAML is valid (no syntax errors)
- Manual: push a test tag (`v0.0.1-rc.1`) to verify workflow triggers and produces a GitHub Release with archives + checksums

## Phase 4: Install Script

**Spec coverage:** DP-14, DP-15, DP-16, DP-17, DP-18

### Tasks

- [x] Create `install.sh` at repo root:
  - Detect OS via `uname -s` → darwin/linux (fail on unsupported)
  - Detect arch via `uname -m` → amd64/arm64 (normalize x86_64→amd64, aarch64→arm64)
  - Default version: fetch latest release tag from GitHub API
  - `VERSION` env var overrides for pinning
  - Download archive + checksums.txt from GitHub Releases
  - Verify SHA256 checksum (shasum -a 256 or sha256sum)
  - Extract binary from archive
  - Install to `INSTALL_DIR` (default `/usr/local/bin`, fallback `~/.local/bin` if not writable)
  - Print success message with installed path and version
- [x] Commit: `feat: add curl|bash install script`

### Success Criteria

- `shellcheck install.sh` passes (if shellcheck available)
- Manual: `bash install.sh` on macOS arm64 installs working binary
- Manual: `VERSION=v0.0.1-rc.1 bash install.sh` installs the pinned version
- Manual: running on unsupported OS prints clear error
