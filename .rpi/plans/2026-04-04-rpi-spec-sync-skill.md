---
date: 2026-04-04T22:54:16+02:00
design: .rpi/designs/2026-04-04-rpi-spec-sync-skill-for-syncing-specs-to-codebase.md
spec: .rpi/specs/spec-sync.md
status: complete
tags:
    - plan
topic: rpi-spec-sync skill
---

# rpi-spec-sync skill — Implementation Plan

## Overview

Create the `/rpi-spec-sync` skill that syncs specs to the codebase. Skill-only — no new Go code, CLI commands, or MCP tools.

**Scope**: 2 new files, 1 modified file

## Source Documents
- **Design**: .rpi/designs/2026-04-04-rpi-spec-sync-skill-for-syncing-specs-to-codebase.md
- **Spec**: .rpi/specs/spec-sync.md

---

## Phase 1: Create Skill and Update Tests

### Overview
Create the `rpi-spec-sync` SKILL.md in both embedded assets and installed location. Add to the pipeline skills test list. Must follow existing conventions: `## Goal`, `## Invariants`, `## Principles` sections, ≤50 lines body, no `rpi_` tool names, no backtick-quoted CLI invocations.

### Tasks:

#### 1. Embedded Skill File
**File**: `internal/workflow/assets/skills/rpi-spec-sync/SKILL.md`
**Changes**: Create new skill with frontmatter (`name: rpi-spec-sync`, `description`), Goal, Invariants, and Principles sections. Body must describe the two-phase workflow (scan → act) and the five actions (keep, rewrite, rename, merge, archive).

#### 2. Installed Skill Copy
**File**: `.claude/skills/rpi-spec-sync/SKILL.md`
**Changes**: Identical to embedded version (installed copies match embedded source).

#### 3. Pipeline Skills Test List
**File**: `internal/workflow/workflow_test.go`
**Changes**: Add `"rpi-spec-sync"` to the `pipelineSkills` slice in `TestPromptStructure_HasRequiredSections` and `TestPromptStructure_LineCount`.

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` passes (structure tests validate sections, line count, no tool references)
- [x] `go build ./cmd/rpi` compiles

#### Manual Verification:
- [x] Review skill content for clarity and completeness against the spec's 8 scenarios

### Commit:
- [x] Stage: `internal/workflow/assets/skills/rpi-spec-sync/SKILL.md`, `internal/workflow/workflow_test.go`, `cmd/rpi/init_cmd_test.go`, `cmd/rpi/update_cmd_test.go`
- [x] Message: `feat: add rpi-spec-sync skill for syncing specs to codebase`

---

## References
- Design: .rpi/designs/2026-04-04-rpi-spec-sync-skill-for-syncing-specs-to-codebase.md
- Spec: .rpi/specs/spec-sync.md
