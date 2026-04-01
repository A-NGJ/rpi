---
archived_date: "2026-04-02"
date: 2026-03-22T23:19:52+01:00
status: archived
tags:
    - design
    - distribution
    - release
topic: Distribution Pipeline
---

# Design: Distribution Pipeline

## Summary

Set up automated cross-platform release builds via goreleaser, a curl|bash installer for frictionless onboarding, and a `rpi version` command. Currently the only install methods are `go install` and `make install` — both require Go toolchain.

## Context

RPI is a Go binary with no runtime dependencies (CGO_ENABLED=0, static linking). The module path is `github.com/A-NGJ/ai-agent-research-plan-implement-flow` and the main package is `cmd/rpi/`. `go install` already works but requires the Go toolchain. Most users of agentic coding tools are not Go developers.

Industry standard for Go CLI distribution: goreleaser for cross-platform builds + GitHub Releases and a curl|bash script for zero-dependency quick install.

## Constraints

- Must produce static binaries (CGO_ENABLED=0) for darwin/linux × amd64/arm64 + windows/amd64
- Must include SHA256 checksums for all artifacts
- Must not require Go toolchain for end-user installation

## Components

### 1. Version Injection

Add build-time version info to `cmd/rpi/main.go` via ldflags:

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

New `rpi version` subcommand outputs `rpi version v0.1.0 (commit abc1234, built 2026-03-22)`. Goreleaser sets these via `-ldflags "-X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}"`.

### 2. Goreleaser Configuration

`.goreleaser.yml` at repo root. Builds for 5 platform/arch combinations (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64). Archives as tar.gz (zip for windows). Auto-generates SHA256 checksums.

### 3. GitHub Actions Release Workflow

`.github/workflows/release.yml` triggered on `v*` tag push. Steps: checkout with full history → setup Go → run goreleaser. Requires `GITHUB_TOKEN` (auto-provided).

Release process: `git tag v0.1.0 && git push origin v0.1.0` → CI builds, creates GitHub Release.

### 4. Install Script

`install.sh` at repo root. Detects OS (darwin/linux/windows) and architecture (amd64/arm64), downloads the correct archive from GitHub Releases, verifies SHA256 checksum, extracts binary, installs to `$INSTALL_DIR` (default `/usr/local/bin`, falls back to `~/.local/bin` if not writable).

Install: `curl -sSfL https://raw.githubusercontent.com/A-NGJ/ai-agent-research-plan-implement-flow/main/install.sh | bash`

Supports `VERSION=v0.1.0` env var for pinning.

## File Structure

**New files:**
- `.goreleaser.yml` — build/release/homebrew config
- `.github/workflows/release.yml` — CI release workflow
- `install.sh` — curl|bash installer
- `cmd/rpi/version.go` — version subcommand

**Modified files:**
- `cmd/rpi/main.go` — add version/commit/date vars

## Risks

- **Install script maintainability** — Archive naming must match goreleaser's `name_template`; if template changes, script breaks. Mitigated by testing in CI.
- **Module path length** — `go install github.com/A-NGJ/ai-agent-research-plan-implement-flow/cmd/rpi@latest` is unwieldy. No fix here (would require module rename), but the install script makes this irrelevant for most users.

## Out of Scope

- Homebrew tap (can add later)
- Code signing with cosign (can add later)
- SBOM generation
- Docker images
- Windows package managers (WinGet, Scoop)
- Module path rename / vanity import

## References

- goreleaser documentation: https://goreleaser.com
- goreleaser GitHub Action: https://github.com/goreleaser/goreleaser-action

