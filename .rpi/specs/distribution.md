---
domain: distribution-pipeline
feature: distribution
last_updated: 2026-04-04T22:30:00+02:00
updated_by: .rpi/designs/2026-03-22-distribution-pipeline.md
---

# Distribution Pipeline

## Purpose

Enable users to install and update the `rpi` binary without a Go toolchain, via curl|bash or GitHub Release downloads with checksum verification.

## Scenarios

### Version command shows dev info for development builds
Given the binary is built with `go build` without ldflags
When the user runs `rpi version`
Then output contains "dev", "none", and "unknown" for version, commit, and date

### Version command shows release info for tagged builds
Given goreleaser builds the binary with ldflags from a git tag
When the user runs `rpi version`
Then output contains the semver tag, short commit hash, and ISO date

### Goreleaser produces cross-platform static binaries
Given the `.goreleaser.yml` configuration exists
When a release build is triggered
Then static binaries (CGO_ENABLED=0) are produced for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, and windows/amd64 with SHA256 checksums

### Install script installs correct binary for OS and architecture
Given a supported OS and architecture (macOS or Linux)
When the user runs the install script
Then the correct binary is downloaded, checksum-verified, and installed to `/usr/local/bin` or `~/.local/bin` as fallback

### Install script verifies checksums before installing
Given a corrupted or tampered archive download
When the install script runs
Then it exits with a non-zero code and prints a checksum mismatch error

### Install script handles unsupported platforms
Given an unsupported OS or architecture
When the install script runs
Then it exits with a clear "Unsupported OS" error message

## Constraints
- All binaries statically linked (CGO_ENABLED=0)
- SHA256 checksums for every artifact
- `go install` continues to work (no breaking module changes)
- Install script works on macOS and Linux without dependencies beyond curl and tar
- Do not require sudo by default

## Out of Scope
- Homebrew tap
- Code signing (cosign) or SBOM generation
- Docker images or Windows package managers
- Module path rename
