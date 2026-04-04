---
date: 2026-04-04T22:20:27+02:00
design: .rpi/designs/2026-04-04-gherkin-inspired-spec-format-and-spec-aware-verification.md
spec: .rpi/specs/spec-format.md
status: complete
tags:
    - plan
topic: gherkin-inspired spec format and spec-aware verification
---

# Gherkin-Inspired Spec Format and Spec-Aware Verification — Implementation Plan

## Overview

Replace XX-N requirement specs with Gherkin-inspired scenario format, add `rpi verify spec` CLI/MCP tool for structured scenario parsing, update skill prompts, and migrate existing specs.

**Scope**: 5 files modified, 0 new files

## Source Documents
- **Design**: .rpi/designs/2026-04-04-gherkin-inspired-spec-format-and-spec-aware-verification.md
- **Spec**: .rpi/specs/spec-format.md

---

## Phase 1: Template and Skill Prompts

### Overview
Update the spec template and both skill prompts to use the Gherkin-inspired scenario format. All text/markdown changes — no Go code.

### Tasks:

#### 1. Spec Template
**File**: `internal/workflow/assets/templates/spec.tmpl`
**Changes**:
- Replace `id` frontmatter field with `feature`
- Replace `## Behavior` section (with XX-N items) with `## Scenarios` section (with `### Scenario title` + Given/When/Then)
- Remove `## Test Cases` section
- Collapse `### Must` / `### Must Not` / `### Out of Scope` into flat `## Constraints` and `## Out of Scope`

#### 2. rpi-propose Skill
**File**: `.claude/skills/rpi-propose/SKILL.md`
**Changes**:
- Line 27: Replace "Create a behavioral spec with prefixed IDs (XX-N), constraints (must/must-not/out-of-scope), and test cases (given/when/then)" with scenario-based invariant
- Add guidance that scenarios must describe user-observable behavior, not internal structure
- Add target of 5-8 scenarios per spec

#### 3. rpi-verify Skill
**File**: `.claude/skills/rpi-verify/SKILL.md`
**Changes**:
- Line 20: Replace "verify spec behaviors (XX-N) against actual code and tests" with "use `rpi verify spec` to extract scenarios, then verify each scenario against actual code with pass/fail per scenario"

### Success Criteria:

#### Automated Verification:
- [x] `rpi scaffold spec --topic "test feature"` outputs markdown with `feature:` in frontmatter (not `id:`)
- [x] `rpi scaffold spec --topic "test feature"` outputs markdown with `## Scenarios` section containing Given/When/Then
- [x] `go build ./cmd/rpi` still compiles (template embedded at build time)

#### Manual Verification:
- [x] Review rpi-propose skill invariant references scenarios, not XX-N
- [x] Review rpi-verify skill invariant references `rpi_verify_spec` MCP tool, not XX-N

### Commit:
- [x] Stage: `internal/workflow/assets/templates/spec.tmpl`, `.rpi/templates/spec.tmpl`, `.claude/skills/rpi-propose/SKILL.md`, `.claude/skills/rpi-verify/SKILL.md`
- [x] Message: `refactor: replace XX-N spec format with Gherkin-inspired scenarios`

---

## Phase 2: Verify Spec CLI Subcommand and MCP Tool

### Overview
Add scenario parsing to `rpi verify`, register as MCP tool, and add tests. This gives both the CLI and Claude structured access to spec scenarios.

### Tasks:

#### 1. Scenario Parser and CLI Action
**File**: `cmd/rpi/verify.go`
**Changes**:
- Add `Scenario` struct: `Title`, `Given`, `When`, `Then` (all strings)
- Add `SpecResult` struct: `Spec` (path), `Feature` (from frontmatter), `Scenarios` ([]Scenario), `Total` (int)
- Add `parseScenarios(content string) []Scenario` — extract `### ` blocks under `## Scenarios`, parse Given/When/Then keywords (handling multi-line continuation)
- Add `runVerifySpec(specPath, format string) error` — read file, parse frontmatter for `feature`, extract `## Scenarios` section via `frontmatter.ExtractSection`, parse scenarios, output JSON
- Add `"spec"` case to `runVerify` switch
- Update command help text and examples to include `spec` action

#### 2. MCP Tool Registration
**File**: `cmd/rpi/serve.go`
**Changes**:
- Add `verifySpecInput` struct with `SpecPath string` field
- Add `handleVerifySpec` handler (same logic as CLI, returns JSON via `jsonResult`)
- Register `rpi_verify_spec` tool in the MCP tool list alongside `rpi_verify_completeness` and `rpi_verify_markers`

#### 3. Tests
**File**: `cmd/rpi/verify_test.go`
**Changes**:
- `TestParseScenarios` — basic spec with 3 scenarios, verify title/given/when/then extraction
- `TestParseScenariosMultiLine` — Given/When/Then with continuation lines (no keyword prefix)
- `TestParseScenariosEmpty` — spec with `## Scenarios` but no scenario blocks → empty slice
- `TestParseScenariosNoSection` — spec without `## Scenarios` → empty slice

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` passes (all new and existing tests)
- [x] `go build ./cmd/rpi` compiles
- [x] `rpi verify spec .rpi/specs/spec-format.md` outputs valid JSON with 8 scenarios

### Commit:
- [x] Stage: `cmd/rpi/verify.go`, `cmd/rpi/verify_test.go`, `cmd/rpi/serve.go`, `cmd/rpi/serve_test.go`, `internal/template/render_test.go`, `internal/workflow/assets/skills/rpi-verify/SKILL.md`, `internal/workflow/assets/skills/rpi-implement/SKILL.md`
- [x] Message: `feat(verify): add spec subcommand for scenario parsing`

---

## Phase 3: Migrate Existing Specs

### Overview
Rewrite all existing XX-N specs to the new scenario format. Each spec gets 5-8 behavioral scenarios preserving domain coverage and constraints. This is the final step that removes dual-format support.

### Tasks:

#### 1. Migrate each spec
**Files**: All `.rpi/specs/*.md` (10 specs, excluding the already-scenario-format `gherkin-inspired-spec-format-and-spec-aware-verification.md`)
- `rpi-explain-command.md` (9 requirements → ~5-6 scenarios)
- `agent-skills-compatibility.md` (14 requirements → ~6-7 scenarios)
- `distribution-pipeline.md` (14 requirements → ~6-7 scenarios)
- `init-update-cleanup.md` (15 requirements → ~6-7 scenarios)
- `index-expansion.md` (26 requirements → ~7-8 scenarios)
- `benchmark-project-for-rpi-flow-quality-testing.md` (20 requirements → ~6-7 scenarios)
- `mcp-init-and-commands.md` (10 requirements → ~5-6 scenarios)
- `rpi-status-dashboard-command.md` (19 requirements → ~6-7 scenarios)
- `unified-status-lifecycle.md` (14 requirements → ~5-6 scenarios)
- `remove-index-subsystem-from-rpi.md` (9 requirements → ~5-6 scenarios)

For each spec:
- Replace `id:` with `feature:` in frontmatter
- Replace `## Behavior` + XX-N lists with `## Scenarios` + Given/When/Then blocks
- Remove `## Test Cases` section
- Collapse constraint subsections into flat `## Constraints` + `## Out of Scope`
- Preserve domain coverage — every key behavior must be represented by a scenario

### Success Criteria:

#### Automated Verification:
- [x] `rpi verify spec` successfully parses each migrated spec (valid JSON, 5-8 scenarios each)
- [x] `go test ./...` passes (status command updated to count scenarios instead of requirements)

#### Manual Verification:
- [x] Review migrated specs for behavioral coverage — no key behaviors lost from original XX-N requirements

### Commit:
- [x] Stage: all modified `.rpi/specs/*.md` files, `cmd/rpi/status.go`, `cmd/rpi/status_test.go`
- [x] Message: `refactor(specs): migrate all specs from XX-N to scenario format`

---

## References
- Design: .rpi/designs/2026-04-04-gherkin-inspired-spec-format-and-spec-aware-verification.md
- Spec: .rpi/specs/spec-format.md
