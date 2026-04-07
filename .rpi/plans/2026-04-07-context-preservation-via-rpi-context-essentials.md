---
date: 2026-04-07T12:29:06+02:00
design: .rpi/designs/2026-04-07-context-preservation-via-rpi-context-essentials.md
spec: .rpi/specs/context-preservation.md
status: complete
tags:
    - plan
topic: context preservation via rpi-context-essentials
---

# Context Preservation via rpi_context_essentials — Implementation Plan

## Overview

Add `rpi context` CLI command and `rpi_context_essentials` MCP tool that returns a compact snapshot of the active implementation context. Wire into Claude Code's PostCompact hook and add prompt guidance to the implement skill.

**Scope**: 2 new files, 4 modified files

## Source Documents
- **Design**: .rpi/designs/2026-04-07-context-preservation-via-rpi-context-essentials.md
- **Spec**: .rpi/specs/context-preservation.md

---

## Phase 1: Core Context Assembly Logic + CLI Command

### Overview
Implement the context assembly engine and CLI command in `cmd/rpi/context.go`. This covers auto-detection of active plans, chain resolution for linked specs/designs, current phase detection, and compact output serialization. All reuses existing internals.

### Tasks:

#### 1. Output Structs
**File**: `cmd/rpi/context.go`
**Changes**:
- Define `ContextResult` struct with nested `PlanContext`, `SpecContext`, `GitContext` fields
- `PlanContext`: `Path`, `Topic`, `CurrentPhase`, `Progress` (checked/total), `NextItems` ([]string, max 3)
- `SpecContext`: `Path`, `Feature`, `ScenarioTitles` ([]string)
- `GitContext`: `Branch`, `UncommittedFiles` (int)
- Top-level `Constraints` field (string, truncated to 200 chars)

#### 2. Context Assembly Function
**File**: `cmd/rpi/context.go`
**Changes**:
- `assembleContext(rpiDir, planPath string) (*ContextResult, error)` — main orchestration function
- When `planPath` is empty: call `scanner.Scan` with `Type: "plan", Status: "active"`, pick the most recently dated plan by parsing `date` frontmatter field. Return empty `ContextResult` if no active plans found.
- When `planPath` is provided: use it directly
- Read plan file, call `parseCheckboxes` to get progress. Detect current phase: walk phases in order, first phase with unchecked items is current. Extract up to 3 next unchecked item texts.
- Resolve chain from plan via `chain.Resolve` with `Sections: ["Constraints"]`. Find linked spec (type "spec") and design (type "design") in chain artifacts.
- For spec: call `parseScenarios` on the spec body, collect titles only.
- For design: use the extracted `Constraints` section from chain resolution, truncate to 200 chars if needed.
- For git: call `git.GatherContext()` for branch, call `git.ChangedFiles()` for uncommitted file count.

#### 3. Current Phase Detection Helper
**File**: `cmd/rpi/context.go`
**Changes**:
- `detectCurrentPhase(content string) (phaseName string, nextItems []string)` — reuses `phaseRe` and `uncheckedRe` from verify.go
- Walks lines: track current phase heading. For each phase, check if it has unchecked items. First phase with unchecked items is the current phase. Collect up to 3 unchecked item texts from that phase.
- If all phases are complete, return empty strings (plan is done).

#### 4. CLI Command
**File**: `cmd/rpi/context.go`
**Changes**:
- `contextCmd` cobra command: `Use: "context [plan-path]"`, `Short: "Show active implementation context"`, `Args: cobra.MaximumNArgs(1)`
- `runContext(cmd, args)`: extract optional plan path from args, call `assembleContext`, marshal to JSON, print
- Register `contextCmd` in `init()`, add `--rpi-dir` flag

#### 5. Tests
**File**: `cmd/rpi/context_test.go`
**Changes**:
- `TestDetectCurrentPhase` — plan with 3 phases, first 2 complete, third has unchecked items → returns phase 3 name + up to 3 next items
- `TestDetectCurrentPhaseAllComplete` — all phases checked → returns empty
- `TestDetectCurrentPhaseMaxItems` — phase with 5 unchecked items → returns only first 3
- `TestAssembleContextAutoDetect` — create temp `.rpi/` with an active plan + linked spec + design, verify auto-detection picks it up and output includes all fields
- `TestAssembleContextExplicitPath` — provide plan path directly, verify it's used regardless of other active plans
- `TestAssembleContextNoPlan` — no active plans → returns empty result with no error
- `TestAssembleContextConstraintsTruncation` — design with >200 char constraints → verify truncation

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestDetectCurrentPhase` passes
- [x] `go test ./cmd/rpi/ -run TestAssembleContext` passes
- [x] `go build ./cmd/rpi` compiles
- [x] `rpi context` runs without error (returns empty result or active plan context)

### Commit:
- [x] Stage: `cmd/rpi/context.go`, `cmd/rpi/context_test.go`
- [x] Message: `feat(context): add rpi context command for implementation context snapshots`

---

## Phase 2: MCP Tool + Hook Wiring + Skill Prompt

### Overview
Register `rpi_context_essentials` as an MCP tool, add PostCompact hook configuration to `rpi init`/`rpi update`, and add context recovery guidance to the implement skill prompt.

### Tasks:

#### 1. MCP Tool Registration
**File**: `cmd/rpi/serve.go`
**Changes**:
- Add `contextInput` struct: `PlanPath string` (optional, jsonschema description)
- Add `handleContext` handler: calls `assembleContext(rpiDirFlag, input.PlanPath)`, returns via `jsonResult`
- Register `rpi_context_essentials` tool in `registerTools` with description derived from `contextCmd`

#### 2. PostCompact Hook Configuration
**File**: `cmd/rpi/init_cmd.go`
**Changes**:
- Add `configureHooks(w io.Writer, toolDirPath string)` function — same merge pattern as `configureSettings`:
  - Read existing `settings.json`
  - Parse `hooks` key if present
  - Check if `PostCompact` already has the RPI entry (idempotent)
  - If not present, add the PostCompact hook entry: `{"type": "command", "command": "cat <<'HOOK_EOF'\nIMPORTANT: Context was compacted. Call the rpi_context_essentials MCP tool to restore your implementation context (active plan phase, spec scenarios, constraints).\nHOOK_EOF"}`
  - Write back with indentation

#### 3. Wire Hook Config into Sync
**File**: `cmd/rpi/sync.go`
**Changes**:
- Call `configureHooks(opts.w, ...)` after `configureSettings` for Claude target

#### 4. Skill Prompt Update
**Files**: `internal/workflow/assets/skills/rpi-implement/SKILL.md`, `.claude/skills/rpi-implement/SKILL.md`
**Changes**:
- Add invariant to the `## Invariants` section: `- **Context recovery**: if context seems lost or you're unsure which phase you're on, call rpi_context_essentials to restore your implementation context`

#### 5. Tests
**File**: `cmd/rpi/init_cmd_test.go`
**Changes**:
- `TestConfigureHooksAddsPostCompact` — call `configureHooks` on a dir with settings.json, verify PostCompact hook is present
- `TestConfigureHooksIdempotent` — call twice, verify only one PostCompact entry
- `TestConfigureHooksMergesExisting` — settings.json with existing hooks, verify RPI hook is added without clobbering

**File**: `cmd/rpi/serve_test.go` (if exists, or inline verification)
**Changes**:
- Verify `rpi_context_essentials` is registered in the tool list

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestConfigureHooks` passes
- [x] `go test ./...` passes (all existing + new tests)
- [x] `go build ./cmd/rpi` compiles

#### Manual Verification:
- [x] Review rpi-implement skill prompt includes context recovery invariant in both copies
- [x] Review PostCompact hook message is clear and actionable

### Commit:
- [x] Stage: `cmd/rpi/serve.go`, `cmd/rpi/init_cmd.go`, `cmd/rpi/init_cmd_test.go`, `cmd/rpi/sync.go`, `internal/workflow/assets/skills/rpi-implement/SKILL.md`, `.claude/skills/rpi-implement/SKILL.md`
- [x] Message: `feat(context): register MCP tool, add PostCompact hook, update implement skill`

---

## References
- Design: .rpi/designs/2026-04-07-context-preservation-via-rpi-context-essentials.md
- Spec: .rpi/specs/context-preservation.md
