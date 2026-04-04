---
archived_date: "2026-04-02"
date: 2026-03-24T00:55:02+01:00
design: .rpi/designs/2026-03-24-init-update-cleanup.md
spec: .rpi/specs/init-update.md
status: archived
tags:
    - plan
topic: init-update-cleanup
---

# init-update-cleanup — Implementation Plan

## Overview

Extract shared sync logic from `init` and `update` into `syncProject()`, fix rules file overwrite inconsistency, update tests.

**Scope**: 1 new file, 2 modified files, 1 modified test file

## Source Documents
- **Design**: .rpi/designs/2026-03-24-init-update-cleanup.md
- **Spec**: .rpi/specs/init-update.md

## Phase 1: Extract syncProject and wire into both commands

### Overview
Create `sync.go` with the shared `syncProject()` function. Refactor `runInit` and `runUpdate` to delegate to it. Rules file now respects `force` uniformly.

### Tasks:

#### 1. Create `cmd/rpi/sync.go`
**File**: `cmd/rpi/sync.go`
**Changes**:
- Define `rpiSubdirs` variable (single source of truth for `.rpi/` subdirectory list) — spec IU-2
- Define `syncOptions` struct with fields: `targetDir`, `cfg` (targetConfig), `force`, `skipRules`, `w` (io.Writer)
- Implement `syncProject(opts syncOptions) error` that:
  1. Ensures `.rpi/` subdirs exist (create only if missing)
  2. Ensures tool subdirs exist if `cfg.toolDir != ""` (create only if missing)
  3. Installs skills via `workflow.InstallSkills(skillsDir, cfg.target, opts.force)` — spec IU-1
  4. Installs templates via `workflow.InstallTemplates(templatesDir, opts.force)` — spec IU-1
  5. Writes rules file: skip if `opts.skipRules` or `cfg.rulesFile == ""`; otherwise write only if file doesn't exist OR `opts.force` — spec IU-3, IU-4, IU-5, IU-6
  6. Calls `configureSettings()` for Claude target — spec IU-1
  7. Rebuilds codebase index — spec IU-1

#### 2. Simplify `runInit`
**File**: `cmd/rpi/init_cmd.go`
**Changes**:
- Remove duplicated `.rpi/` subdir creation, skill install, template install, rules file generation, and index build logic
- Remove the `rpiSubdirs` slice (now in `sync.go`)
- Keep: target resolution, guard (tool dir exists check), tool dir creation, `.gitignore` management, MCP registration — spec IU-7, IU-8, IU-9, IU-10
- After init-only setup, call `syncProject(syncOptions{targetDir, cfg, force: false, skipRules: initNoClaudeMD, w})`

#### 3. Simplify `runUpdate`
**File**: `cmd/rpi/update_cmd.go`
**Changes**:
- Remove duplicated `.rpi/` subdir creation, tool subdir creation, skill install, template install, rules file update, settings.json, and index build logic
- Keep: `.rpi/` existence guard, `detectTarget()` — spec IU-11, IU-12
- Call `syncProject(syncOptions{targetDir, cfg, force: updateForce, skipRules: updateNoClaudeMD, w})`

### Success Criteria:

#### Automated Verification:
- [x] `go build ./cmd/rpi/` succeeds
- [x] `go vet ./cmd/rpi/` clean
- [x] All existing tests pass except `TestUpdateUpdatesRulesFile` and `TestUpdateRegeneratesIndex` (expected — behavior/message changes, fixed in Phase 2)

### Commit:
- [x] Stage: `cmd/rpi/sync.go`, `cmd/rpi/init_cmd.go`, `cmd/rpi/update_cmd.go`
- [x] Message: `refactor: extract syncProject from init/update commands`

---

## Phase 2: Update tests for rules file behavior change

### Overview
Fix the one test that asserts the old rules-file-always-overwrite behavior, add a new test for `--force` overwriting the rules file. Verify full suite green.

### Tasks:

#### 1. Update rules file tests
**File**: `cmd/rpi/update_cmd_test.go`
**Changes**:
- **Modify `TestUpdateUpdatesRulesFile`**: Assert that `update` without `--force` does NOT overwrite a customized CLAUDE.md — spec IU-3
- **Add `TestUpdateForceOverwritesRulesFile`**: Init, customize CLAUDE.md, run `update --force`, assert CLAUDE.md is replaced with template — spec IU-4
- Existing `TestUpdateNoClaudeMDSkipsRulesFile` already covers IU-5 — verify it still passes

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -count=1` — all 116 tests pass (0 failures)
- [x] `go vet ./...` clean

### Commit:
- [x] Stage: `cmd/rpi/update_cmd_test.go`
- [x] Message: `test: update rules file tests for --force consistency`

---

## References
- Design: .rpi/designs/2026-03-24-init-update-cleanup.md
- Spec: .rpi/specs/init-update.md
