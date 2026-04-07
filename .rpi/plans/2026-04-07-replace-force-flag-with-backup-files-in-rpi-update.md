---
date: 2026-04-07T18:19:09+02:00
spec: .rpi/specs/init-update.md
status: complete
tags:
    - plan
topic: replace force flag with backup files in rpi update
---

# Replace --force Flag with Backup Files in rpi update

## Overview

Remove the `--force` flag from `rpi update`. Instead, `rpi update` always overwrites managed files with the latest embedded versions, creating `.bak` backups of any files that differ before overwriting.

**Scope**: 5 files modified (workflow.go, sync.go, update_cmd.go, init_cmd.go, update_cmd_test.go), 1 spec updated

## Source Documents
- **Spec**: .rpi/specs/init-update.md

## Phase 1: Add backup-before-write logic to workflow Install functions

### Overview
Replace the `force bool` parameter with automatic backup behavior: always write, but create a `.bak` copy if the existing file has different content.

### Tasks:

#### 1. workflow.go — InstallSkills
**File**: `internal/workflow/workflow.go`
**Changes**:
- Remove `force bool` parameter from `InstallSkills`
- Before writing, if dest file exists and content differs from embedded, copy existing file to `dest + ".bak"`
- Always write the file (skip only if content is identical)
- Return `(installed int, backedUp int, err error)` instead of `(int, error)`

#### 2. workflow.go — InstallTemplates
**File**: `internal/workflow/workflow.go`
**Changes**: Same pattern as InstallSkills — remove `force`, add backup logic, return `(installed, backedUp, error)`

#### 3. workflow.go — InstallAgents
**File**: `internal/workflow/workflow.go`
**Changes**: Same pattern as above

### Success Criteria:

#### Automated Verification:
- [x] `go build ./...` compiles (will fail until Phase 2 updates callers — verify in Phase 2)

### Commit:
- [x] Stage: `internal/workflow/workflow.go`
- [x] Message: `refactor(workflow): replace force parameter with automatic backup-before-write`

---

## Phase 2: Update sync.go, update_cmd.go, and init_cmd.go

### Overview
Remove `force` from `syncOptions` and all call sites. Update rules file and hooks logic to always write with backup.

### Tasks:

#### 1. sync.go — remove force from syncOptions and update syncProject
**File**: `cmd/rpi/sync.go`
**Changes**:
- Remove `force` field from `syncOptions`
- Update `InstallSkills`, `InstallTemplates`, `InstallAgents` calls to match new signatures
- Log backup counts (e.g., "Backed up 3 modified files")
- Rules file: always write; if existing content differs, write `.bak` first
- Update `configureHooks` call to remove `force` argument

#### 2. update_cmd.go — remove --force flag
**File**: `cmd/rpi/update_cmd.go`
**Changes**:
- Remove `updateForce` variable
- Remove `--force` flag registration from `init()`
- Remove `force` from `syncProject` call in `runUpdate`
- Update command Long description and Examples to remove --force references

#### 3. init_cmd.go — remove force from syncProject call and configureHooks
**File**: `cmd/rpi/init_cmd.go`
**Changes**:
- Remove `force: true` from `syncProject` call
- Remove `force` parameter from `configureHooks` function signature
- Remove force guard inside `configureHooks` — always replace RPI hook entries

### Success Criteria:

#### Automated Verification:
- [x] `go build ./...` compiles cleanly
- [x] `go vet ./...` passes (test files fixed in Phase 3)

### Commit:
- [x] Stage: `cmd/rpi/sync.go`, `cmd/rpi/update_cmd.go`, `cmd/rpi/init_cmd.go`
- [x] Message: `refactor(update): remove --force flag, always sync with backups`

---

## Phase 3: Update tests

### Overview
Rewrite force-related tests to verify the new backup behavior.

### Tasks:

#### 1. Rewrite force-related tests
**File**: `cmd/rpi/update_cmd_test.go`
**Changes**:
- Remove `resetUpdateFlags` references to `updateForce`
- `TestUpdateDoesNotOverwriteWithoutForce` → rename to `TestUpdateBacksUpModifiedFiles`: verify that when a skill file has custom content, update overwrites it AND creates a `.bak` file with the old content
- `TestUpdateForceOverwritesFiles` → remove (covered by above)
- `TestUpdatePreservesRulesFileWithoutForce` → rename to `TestUpdateBacksUpModifiedRulesFile`: verify update writes new rules file AND creates `.bak` with old content
- `TestUpdateForceOverwritesRulesFile` → remove (covered by above)
- Add `TestUpdateSkipsIdenticalFiles`: verify that when file content matches embedded, no `.bak` is created and file is not rewritten
- Update `TestUpdatePreservesExistingCommandsDir` to remove `updateForce = true` (update always overwrites now)

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestUpdate` — all tests pass
- [x] `go test ./...` — full suite passes

### Commit:
- [x] Stage: `cmd/rpi/update_cmd_test.go`
- [x] Message: `test(update): rewrite tests for backup-on-update behavior`

---

## Phase 4: Update spec

### Overview
Update the init-update spec to reflect the new backup behavior.

### Tasks:

#### 1. Update spec scenarios
**File**: `.rpi/specs/init-update.md`
**Changes**:
- Remove "Update without --force preserves rules file" scenario
- Remove "Update with --force overwrites rules file" scenario
- Add "Update backs up modified files before overwriting" scenario
- Add "Update skips backup when file content is identical" scenario
- Update Constraints to remove `--force` references

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` — still passes (spec is documentation, not code)

#### Manual Verification:
- [x] Spec scenarios accurately describe the new behavior

### Commit:
- [x] Stage: `.rpi/specs/init-update.md`
- [x] Message: `docs(spec): update init-update spec for backup-on-update behavior`

---

## References
- Spec: .rpi/specs/init-update.md
- Key files: `internal/workflow/workflow.go`, `cmd/rpi/sync.go`, `cmd/rpi/update_cmd.go`, `cmd/rpi/init_cmd.go`
