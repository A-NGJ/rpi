---
archived_date: "2026-04-02"
date: 2026-03-22T22:18:51+01:00
design: .rpi/designs/2026-03-22-rpi-diagnose-command.md
spec: .rpi/specs/2026-03-22-rpi-diagnose-command.md
status: archived
tags:
    - plan
topic: rpi-diagnose command
---

# rpi-diagnose command — Implementation Plan

## Overview

Add the `/rpi-diagnose` command for iterative bug analysis. This is a purely additive change: one new embedded command file and one template update.

**Scope**: 1 new file, 1 modified file

## Source Documents
- **Design**: .rpi/designs/2026-03-22-rpi-diagnose-command.md
- **Spec**: .rpi/specs/2026-03-22-rpi-diagnose-command.md

## Phase 1: Create the embedded command file

### Overview

Create `internal/workflow/assets/commands/rpi-diagnose.md` following the established command pattern (frontmatter, Goal, Invariants, Principles). Encode all spec behaviors DG-1 through DG-9 as invariants.

### Tasks:

#### 1. Command definition
**File**: `internal/workflow/assets/commands/rpi-diagnose.md` (new)
**Changes**:
- Frontmatter: `description`, `model: inherit`, `disable-model-invocation: true`
- Goal section: describe the command's purpose and pipeline position (research-like intake + implement-like fix loop)
- Invariants section: encode DG-1 (intake clarification), DG-2 (reproduction), DG-3 (root cause tracing), DG-4 (max 3 autonomous fix attempts), DG-5 (user checkpoint after 3 failures), DG-6 (commit on success), DG-7 (diagnosis artifact), DG-8 (escalation path), DG-9 (revert on failure)
- Principles section: operational guidance (SDD framing, right-size iteration, escalate don't force)

### Success Criteria:

#### Automated Verification:
- [x] `go build ./...` passes
- [x] `go test ./internal/workflow/...` passes
- [x] File exists at `internal/workflow/assets/commands/rpi-diagnose.md`

#### Manual Verification:
- [x] Each DG-N behavior from spec maps to at least one invariant in the command

### Commit:
- [x] Stage: `internal/workflow/assets/commands/rpi-diagnose.md`
- [x] Message: `feat: add /rpi-diagnose command for iterative bug analysis`

---

## Phase 2: Update CLAUDE.md.template

### Overview

Update the CLAUDE.md template to reference `/rpi-diagnose` and the `diagnoses/` directory so that new projects initialized with `rpi init` include the command in their documentation.

### Tasks:

#### 1. Template update
**File**: `internal/workflow/assets/templates/CLAUDE.md.template` (modify)
**Changes**:
- Add `├── diagnoses/    # Bug diagnosis post-mortems (created by /rpi-diagnose)` to the `.rpi/` directory tree
- Add `- **Diagnose** (\`/rpi-diagnose\`): Iterative root-cause analysis and fix for complex bugs. Optional.` to the pipeline list
- Update the "start with" guidance: add `/rpi-diagnose` for complex bugs

### Success Criteria:

#### Automated Verification:
- [x] `go build ./...` passes
- [x] `go test ./...` passes

#### Manual Verification:
- [x] Template directory tree includes `diagnoses/`
- [x] Template pipeline list includes `/rpi-diagnose`

### Commit:
- [x] Stage: `internal/workflow/assets/templates/CLAUDE.md.template`
- [x] Message: `docs: add /rpi-diagnose to CLAUDE.md template`

---

## Phase 3: Verify spec coverage

### Overview

Walk through each DG-N behavior in the spec and confirm it is represented in the command invariants. No code changes — verification only.

### Tasks:

#### 1. Spec-to-invariant mapping review
**Changes**: Verify each behavior ID maps to command invariants:
- DG-1 (intake clarification) → invariant about interviewing before investigating
- DG-2 (reproduction) → invariant about reproducing before fixing
- DG-3 (root cause tracing) → invariant about file:line evidence
- DG-4 (autonomous loop) → invariant about max 3 attempts
- DG-5 (user checkpoint) → invariant about stopping after 3 failures
- DG-6 (commit on success) → invariant about commit approval flow
- DG-7 (diagnosis artifact) → invariant about producing artifact
- DG-8 (escalation path) → invariant about suggesting /rpi-plan or /rpi-propose
- DG-9 (revert on failure) → invariant about reverting between attempts

### Success Criteria:

#### Manual Verification:
- [x] All 9 DG-N behaviors have corresponding invariants
- [x] No spec behaviors were silently dropped

---

## References
- Design: .rpi/designs/2026-03-22-rpi-diagnose-command.md
- Spec: .rpi/specs/2026-03-22-rpi-diagnose-command.md
