---
archived_date: "2026-04-04"
date: 2026-04-04T19:51:05+02:00
design: .rpi/designs/2026-04-04-unified-status-lifecycle-remove-approved-implemented-add-reopen.md
spec: .rpi/specs/status-lifecycle.md
status: archived
tags:
    - plan
topic: unified status lifecycle remove approved implemented add reopen
---

# Unified Status Lifecycle — Implementation Plan

## Overview

Collapse the two parallel status pipelines into one universal lifecycle. Remove `approved` and `implemented` statuses, add `complete → active` reopen transition, and make specs statusless living documents.

**Scope**: 10 files modified, 0 new files

## Source Documents
- **Design**: .rpi/designs/2026-04-04-unified-status-lifecycle-remove-approved-implemented-add-reopen.md
- **Spec**: .rpi/specs/status-lifecycle.md

## Phase 1: State machine + transition tests (UL-1 through UL-6)

### Overview
Rewrite the state machine to remove `approved`/`implemented` and add `complete → active` reopen. Update all transition tests.

### Tasks:

#### 1. Rewrite state machine
**File**: `internal/frontmatter/transition.go`
**Changes**:
- Replace `validTransitions` map with:
  ```
  draft    → active, superseded
  active   → complete, superseded
  complete → active, archived, superseded
  ```
- Remove `approved` and `implemented` entries entirely

#### 2. Rewrite transition tests
**File**: `internal/frontmatter/frontmatter_test.go`
**Changes**:
- `TestTransitionValid`: Replace cases with new valid set — `draft→active`, `draft→superseded`, `active→complete`, `active→superseded`, `complete→active`, `complete→archived`, `complete→superseded`
- `TestTransitionInvalid`: Replace cases — include `draft→approved`, `active→implemented`, `draft→complete`, `draft→archived`, `active→draft`, `active→archived`, `complete→draft`
- `TestTransitionFromArchived`: Keep as-is (UL-4)
- `TestTransitionMissingStatus`: Keep as-is (UL-5)

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/frontmatter/ -run TestTransition` passes
- [x] `go test ./internal/frontmatter/` passes (all tests green)
- [x] `go vet ./internal/frontmatter/` clean

### Commit:
- [ ] Stage: `internal/frontmatter/transition.go`, `internal/frontmatter/frontmatter_test.go`
- [ ] Message: `refactor(frontmatter): unify status lifecycle, remove approved/implemented, add reopen`

---

## Phase 2: Archivable filter + scanner tests (UL-8, UL-11, UL-12)

### Overview
Update the archivable filter so specs are never surfaced and `implemented` is no longer a valid archivable status.

### Tasks:

#### 1. Update archivable filter
**File**: `internal/scanner/scan.go` (lines 118-129)
**Changes**:
- Specs: return `false` unconditionally (regardless of status)
- Non-specs: archivable when `complete` or `superseded` only (remove `implemented`)

#### 2. Update scanner tests
**File**: `internal/scanner/scan_test.go`
**Changes**:
- Line ~37: Change spec fixture from `status: implemented` to no status field (or remove status line)
- Line ~152: Update archivable test comment — specs are never archivable via scanner (not "only when superseded")

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/scanner/` passes
- [x] `go vet ./internal/scanner/` clean

### Commit:
- [ ] Stage: `internal/scanner/scan.go`, `internal/scanner/scan_test.go`
- [ ] Message: `refactor(scanner): specs never archivable, remove implemented from archivable filter`

---

## Phase 3: Status display + chain test (UL-13, UL-14)

### Overview
Remove `implemented` from status display order and fix chain test fixture using `approved`.

### Tasks:

#### 1. Update status display order
**File**: `cmd/rpi/status.go` (line 63)
**Changes**:
- Remove `"implemented"` from `statusDisplayOrder` slice

#### 2. Fix chain test fixture
**File**: `internal/chain/resolve_test.go` (line 277)
**Changes**:
- Change `status: approved` to remove status field (specs are statusless)

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestStatus` passes
- [x] `go test ./internal/chain/` passes
- [x] `go vet ./cmd/rpi/ ./internal/chain/` clean

### Commit:
- [ ] Stage: `cmd/rpi/status.go`, `internal/chain/resolve_test.go`
- [ ] Message: `refactor(status): remove implemented from display order, fix chain test fixture`

---

## Phase 4: Templates + skills (UL-7, UL-10)

### Overview
Remove `status` field from spec templates and update skill files to stop transitioning spec status.

### Tasks:

#### 1. Update spec templates
**Files**: `.rpi/templates/spec.tmpl`, `internal/workflow/assets/templates/spec.tmpl`
**Changes**:
- Remove the `status: draft` line from frontmatter in both files

#### 2. Update rpi-propose skill
**File**: `.claude/skills/rpi-propose/SKILL.md`
**Changes**:
- Line 29: `"Transition artifacts: design → active, spec → approved, research → complete"` → `"Transition artifacts: design → active, research → complete (if fully addressed)"`
- Line 37: `"Specs are the contract — every design culminates in an approved spec"` → `"Specs are the contract — every design culminates in a spec"`

#### 3. Update rpi-implement skill
**File**: `.claude/skills/rpi-implement/SKILL.md`
**Changes**:
- Line 10: `"Execute an approved plan"` → `"Execute an active plan"`
- Line 30: `"transition spec → implemented, plan → complete"` → `"plan → complete"`

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/workflow/...` passes (template rendering tests)
- [x] `go vet ./...` clean

#### Manual Verification:
- [x] Verify `rpi scaffold spec --topic test` output has no `status` field

### Commit:
- [ ] Stage: `.rpi/templates/spec.tmpl`, `internal/workflow/assets/templates/spec.tmpl`, `.claude/skills/rpi-propose/SKILL.md`, `.claude/skills/rpi-implement/SKILL.md`
- [ ] Message: `refactor(templates,skills): specs are statusless, remove legacy status references`

---

## References
- Design: .rpi/designs/2026-04-04-unified-status-lifecycle-remove-approved-implemented-add-reopen.md
- Spec: .rpi/specs/status-lifecycle.md
