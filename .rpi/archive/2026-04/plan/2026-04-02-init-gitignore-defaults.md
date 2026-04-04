---
archived_date: "2026-04-04"
date: 2026-04-02T01:21:35+02:00
status: archived
tags:
    - plan
topic: init-gitignore-defaults
---

# init-gitignore-defaults — Implementation Plan

## Overview

Flip the `.rpi/` gitignore default so artifacts are tracked by default. Replace `--track-rpi` with `--no-track` (opt out). Always gitignore `.rpi/index.json` since it's a generated file.

**Scope**: 4 files modified

## Phase 1: Flip flag and gitignore logic

### Overview
Replace `--track-rpi` with `--no-track`, invert the default, and always add `.rpi/index.json` to `.gitignore`.

### Tasks:

#### 1. Replace flag and gitignore logic
**File**: `cmd/rpi/init_cmd.go`
**Changes**:
- Rename `initTrackRpi` var to `initNoTrack`
- Replace `--track-rpi` flag with `--no-track` (default: `false`)
- Flip the gitignore condition: only add `.rpi/` to `.gitignore` when `--no-track` is set
- Always add `.rpi/index.json` to `.gitignore` (regardless of `--no-track`)
- Update the command's `Long` description to reflect the new flag

#### 2. Update tests
**File**: `cmd/rpi/init_cmd_test.go`
**Changes**:
- Update `resetInitFlags()` to reset `initNoTrack` instead of `initTrackRpi`
- Rename `TestInitTrackRpi` → `TestInitNoTrack`: set `initNoTrack = true`, assert `.rpi/` IS in `.gitignore`
- Update `TestInitCreatesAllDirs`: assert `.rpi/` is NOT in `.gitignore` (new default), assert `.rpi/index.json` IS in `.gitignore`
- Update `TestInitGitignore`: same — `.rpi/` absent, `.rpi/index.json` present
- Update `TestInitAddsRpiToGitignore` → rename to `TestInitAddsIndexJsonToGitignore`: assert `.rpi/index.json` is in `.gitignore`
- Update `TestInitOpenCode`: adjust `.gitignore` assertions if present

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/...` passes
- [x] `go build ./cmd/rpi` succeeds

### Commit:
- [x] Stage: `cmd/rpi/init_cmd.go`, `cmd/rpi/init_cmd_test.go`
- [x] Message: `feat(init): flip .rpi/ to tracked by default, add --no-track flag`

**Note**: If all success criteria are covered by automated checks and they pass, proceed to the next phase. Only pause for manual confirmation when the phase includes manual verification items.

---

## Phase 2: Update spec and README

### Overview
Update the init-update-cleanup spec (IU-8) and README to reflect the new defaults.

### Tasks:

#### 1. Update spec
**File**: `.rpi/specs/init-update.md`
**Changes**:
- Update IU-8: `rpi init` adds `.rpi/index.json` and the tool directory to `.gitignore` (not `.rpi/`)
- Add IU-8a: `rpi init --no-track` also adds `.rpi/` to `.gitignore`

#### 2. Update README
**File**: `README.md`
**Changes**:
- Update the "Initialize your project" section: mention that `.rpi/` is tracked by default, `index.json` is gitignored
- Update the bullet that says "gitignored by default" to reflect new behavior

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/...` still passes (no regressions)

### Commit:
- [x] Stage: `.rpi/specs/init-update.md`, `README.md`
- [x] Message: `docs: update spec and README for new gitignore defaults`

---

## References
- Spec: `.rpi/specs/init-update.md` (IU-8)
- Current implementation: `cmd/rpi/init_cmd.go:174-179`
