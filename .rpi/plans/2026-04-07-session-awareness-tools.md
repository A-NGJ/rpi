---
date: 2026-04-07T12:52:59+02:00
design: .rpi/designs/2026-04-07-session-awareness-tools.md
spec: .rpi/specs/session-awareness.md
status: complete
tags:
    - plan
    - mcp
    - session
    - pipeline
topic: session awareness tools
---

# Session Awareness Tools — Implementation Plan

## Overview

Add two MCP tools (`rpi_session_resume`, `rpi_suggest_next`) and CLI commands (`rpi resume`, `rpi next`) that give AI assistants awareness of active work and pipeline flow. Wire into Claude Code hooks (SessionStart, Stop) for automatic triggering.

**Scope**: 4 new files, 2 modified files

## Source Documents
- **Design**: .rpi/designs/2026-04-07-session-awareness-tools.md
- **Spec**: .rpi/specs/session-awareness.md

## Phase 1: Pipeline Suggestion Engine

### Overview
Implement `suggestNext` — the core pipeline logic that analyzes artifact state and recommends the next action. Includes output structs, pipeline rules, downstream detection, and the `rpi next` CLI command.

**Spec scenarios covered**: #5 (implementation continuation), #6 (verification suggestion), #7 (next pipeline step)

### Tasks:

#### 1. Suggestion Logic + CLI Command
**File**: `cmd/rpi/suggest.go`
**Changes**:
- `Suggestion` output struct with `Action`, `Reasoning`, `Artifact` fields
- `suggestNext(rpiDir, artifactPath string) (*Suggestion, error)` — main entry point
- Internal helpers:
  - `buildDownstreamMaps(artifacts []scanner.ArtifactInfo) (planDesigns, designResearch map[string]bool)` — reads frontmatter for plans (`design` field) and designs (`related_research` field) to build lookup sets
  - Filter/sort helpers for artifacts by type, status, date
- Pipeline rules evaluated in priority order (active plans → active designs without plans → draft designs → complete research without designs → draft research → nothing)
- `rpi next` cobra command registered on `rootCmd`, optional positional arg for `artifact_path`
- Reuses: `scanner.Scan`, `parseCheckboxes`, `chain.Resolve`, `frontmatter.Parse`, `findActivePlan` (from context.go — already exported-ready), `parseDate` (from context.go)

#### 2. Tests
**File**: `cmd/rpi/suggest_test.go`
**Changes**:
- `TestSuggestNextActivePlanUnchecked` — active plan with unchecked items → suggests `/rpi-implement`
- `TestSuggestNextActivePlanAllChecked` — active plan fully checked, linked spec → suggests `/rpi-verify`
- `TestSuggestNextActiveDesignNoPlan` — active design with no downstream plan → suggests `/rpi-plan`
- `TestSuggestNextDraftDesign` — draft design → suggests review
- `TestSuggestNextCompleteResearchNoDesign` — complete research, no downstream design → suggests `/rpi-propose`
- `TestSuggestNextDraftResearch` — draft research → suggests review
- `TestSuggestNextEmpty` — no artifacts → suggests starting new work
- `TestSuggestNextPriorityOrder` — active plan + complete research → plan wins (later stage priority)
- `TestSuggestNextMostRecentWins` — two active plans → most recently dated wins
- `TestBuildDownstreamMaps` — verifies frontmatter-based downstream detection

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestSuggest` — all suggestion tests pass
- [x] `go build ./cmd/rpi/` — compiles cleanly
- [x] `go vet ./cmd/rpi/` — no issues

### Commit:
- [x] Stage: `cmd/rpi/suggest.go`, `cmd/rpi/suggest_test.go`
- [x] Message: `feat(session): add pipeline suggestion engine with rpi next command`

---

## Phase 2: Session Resume

### Overview
Implement `assembleResume` — combines artifact scanning, plan context, and pipeline suggestion into a session-level overview. Includes the `rpi resume` CLI command.

**Spec scenarios covered**: #1 (active artifacts), #2 (plan progress), #3 (suggestion included), #4 (empty state)

### Tasks:

#### 1. Resume Logic + CLI Command
**File**: `cmd/rpi/resume.go`
**Changes**:
- `ResumeResult` output struct with `Artifacts`, `ActivePlan`, `Suggestion` fields
- `ActivePlanSummary` struct — path, topic, current_phase, progress, next_items (subset of `PlanContext` from context.go)
- `assembleResume(rpiDir string) (*ResumeResult, error)` — main entry point:
  1. `scanner.Scan(rpiDir, Filters{})` — all non-archived artifacts
  2. Filter to active/draft for the artifacts list
  3. If active plan exists, build plan context reusing `detectCurrentPhase` + `parseCheckboxes` from context.go
  4. Call `suggestNext` from Phase 1 for the suggestion
  5. Compose result
- `rpi resume` cobra command registered on `rootCmd`, no args
- Reuses: `scanner.Scan`, `detectCurrentPhase`, `parseCheckboxes`, `parseDate`, `findActivePlan`, `suggestNext`

#### 2. Tests
**File**: `cmd/rpi/resume_test.go`
**Changes**:
- `TestResumeActiveArtifacts` — active plan + active design → both appear in artifacts list with type, status, topic
- `TestResumePlanProgress` — active plan with mixed checkboxes → `active_plan` includes current phase, progress, next items
- `TestResumeSuggestionIncluded` — active artifacts → suggestion field is non-nil with action and reasoning
- `TestResumeEmpty` — no artifacts → empty list, nil plan, "start new work" suggestion
- `TestResumeExcludesArchived` — archived artifacts not in the list
- `TestResumeMostRecentPlan` — two active plans → plan context uses most recent

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestResume` — all resume tests pass
- [x] `go test ./cmd/rpi/ -run TestSuggest` — suggestion tests still pass
- [x] `go build ./cmd/rpi/` — compiles cleanly

### Commit:
- [x] Stage: `cmd/rpi/resume.go`, `cmd/rpi/resume_test.go`
- [x] Message: `feat(session): add session resume with rpi resume command`

---

## Phase 3: MCP + Hook Wiring

### Overview
Register both tools in the MCP server and generalize `configureHooks` to add SessionStart and Stop hooks alongside the existing PostCompact hook.

### Tasks:

#### 1. MCP Tool Registration
**File**: `cmd/rpi/serve.go`
**Changes**:
- Add `suggestInput` struct: `ArtifactPath string` with jsonschema tag
- Add `handleSuggestNext` handler calling `suggestNext(rpiDirFlag, input.ArtifactPath)`
- Add `handleSessionResume` handler calling `assembleResume(rpiDirFlag)`
- Register `rpi_suggest_next` and `rpi_session_resume` tools in `registerTools` with descriptions from cobra commands

#### 2. Hook Configuration
**File**: `cmd/rpi/init_cmd.go`
**Changes**:
- Define a `hookDef` struct/table: `{event, marker, command}` entries for PostCompact, SessionStart, Stop
- Refactor `configureHooks` to iterate over the table instead of hardcoding PostCompact
- SessionStart hook command: `cat <<'HOOK_EOF'\nCall the rpi_session_resume MCP tool to see active work and suggested next steps.\nHOOK_EOF`
- Stop hook command: `cat <<'HOOK_EOF'\nCall the rpi_suggest_next MCP tool to determine the appropriate next pipeline step.\nHOOK_EOF`
- Markers for dedup: `"rpi_session_resume"`, `"rpi_suggest_next"`, `"rpi_context_essentials"` (existing)

#### 3. Tests
**File**: `cmd/rpi/init_cmd_test.go`
**Changes**:
- Extend `TestConfigureHooksAddsPostCompact` → `TestConfigureHooksAddsAllHooks` — verify SessionStart, Stop, and PostCompact are all present
- `TestConfigureHooksPreservesExisting` — existing hooks from other sources are not removed
- `TestConfigureHooksIdempotent` — running twice doesn't duplicate entries

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestConfigureHooks` — hook tests pass
- [x] `go test ./cmd/rpi/` — full test suite passes
- [x] `go build ./cmd/rpi/` — compiles cleanly
- [x] `go vet ./cmd/rpi/` — no issues

### Commit:
- [x] Stage: `cmd/rpi/serve.go`, `cmd/rpi/init_cmd.go`, `cmd/rpi/init_cmd_test.go`
- [x] Message: `feat(session): register MCP tools and add SessionStart/Stop hooks`

---

## References
- Design: .rpi/designs/2026-04-07-session-awareness-tools.md
- Spec: .rpi/specs/session-awareness.md
