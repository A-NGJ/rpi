---
domain: DP
id: DP
status: draft
last_updated: 2026-03-22T23:19:53+01:00
updated_by: .rpi/designs/2026-03-22-distribution-pipeline.md
---

# Distribution Pipeline

## Purpose

Enable users to install and update the `rpi` binary without a Go toolchain, via curl|bash or GitHub Release downloads with checksum verification.

## Behavior

### Version Command
- **DP-1**: `rpi version` MUST output version, commit hash, and build date
- **DP-2**: Development builds MUST show `version=dev, commit=none, date=unknown`
- **DP-3**: Release builds MUST show the git tag version, short commit, and ISO date

### Goreleaser Builds
- **DP-4**: `goreleaser release` MUST produce static binaries (CGO_ENABLED=0) for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
- **DP-5**: Each release MUST include a `checksums.txt` with SHA256 hashes for all archives
- **DP-6**: Archives MUST be tar.gz for darwin/linux and zip for windows
- **DP-7**: Archive naming MUST follow `rpi_<version>_<os>_<arch>.<ext>`

### GitHub Actions
- **DP-8**: Pushing a `v*` tag MUST trigger the release workflow
- **DP-9**: The workflow MUST create a GitHub Release with all archives and checksums attached
### Install Script
- **DP-14**: `curl -sSfL <url>/install.sh | bash` MUST install the correct binary for the detected OS and architecture
- **DP-15**: The install script MUST verify SHA256 checksums before installing
- **DP-16**: The install script MUST fail with a clear error on unsupported OS/architecture
- **DP-17**: `VERSION=v0.1.0` env var MUST pin the installed version
- **DP-18**: The install script MUST default to `/usr/local/bin` and fall back to `~/.local/bin` if not writable (without sudo)

## Constraints

### Must
- All binaries statically linked (CGO_ENABLED=0)
- SHA256 checksums for every artifact
- `go install` continues to work (no breaking module changes)
- Install script works on macOS and Linux without dependencies beyond curl and tar

### Must Not
- Require Go toolchain for end-user installation
- Require sudo by default (install to user-writable path as fallback)
- Include platform combinations without CI testing

### Out of Scope
- Homebrew tap
- Code signing (cosign)
- SBOM generation
- Docker images
- Windows package managers
- Module path rename

## Test Cases

### TC-1: Version command — dev build
- **Given** `go build ./cmd/rpi` without ldflags **When** `rpi version` **Then** output contains "dev" and "none"

### TC-2: Version command — release build
- **Given** goreleaser builds with ldflags **When** `rpi version` **Then** output contains semver, short commit hash, and date

### TC-3: Goreleaser dry run
- **Given** `.goreleaser.yml` exists **When** `goreleaser check` **Then** config validates without errors

### TC-4: Archive naming
- **Given** goreleaser build output **When** listing archives **Then** all match `rpi_<semver>_<os>_<arch>.(tar.gz|zip)`

### TC-5: Install script — macOS arm64
- **Given** macOS arm64 system **When** `bash install.sh` **Then** `rpi version` succeeds AND binary is in install dir

### TC-6: Install script — checksum mismatch
- **Given** corrupted archive download **When** `bash install.sh` **Then** script exits with non-zero and prints "Checksum mismatch"

### TC-7: Install script — unsupported platform
- **Given** unsupported OS (e.g., FreeBSD) **When** `bash install.sh` **Then** script exits with "Unsupported OS"
