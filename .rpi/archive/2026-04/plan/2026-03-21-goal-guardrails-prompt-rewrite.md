---
archived_date: "2026-04-02"
date: 2026-03-21T22:57:50+01:00
design: .rpi/designs/2026-03-21-goal-guardrails-prompt-rewrite.md
spec: .rpi/specs/goal-guardrails-prompt-rewrite.md
status: archived
tags:
    - plan
topic: goal-guardrails-prompt-rewrite
---

# Goal+Guardrails Prompt Rewrite — Implementation Plan

## Overview

Replace one-liner MCP tool descriptions with content generated from Cobra `Long`+`Example` fields, then rewrite all 7 command prompts from procedural step-by-step to a compact goal/invariants/principles structure.

**Scope**: 9 files modified, 0 new files

## Source Documents
- **Design**: .rpi/designs/2026-03-21-goal-guardrails-prompt-rewrite.md
- **Spec**: .rpi/specs/goal-guardrails-prompt-rewrite.md

---

## Phase 1: MCP Description Enrichment

### Overview
Add an `mcpDescription()` helper function and replace all 20 hardcoded one-liner descriptions in `registerTools()` with calls that derive content from Cobra commands. Addresses spec behaviors GG-1, GG-2, GG-3, GG-4.

### Tasks:

#### 1. Add `mcpDescription()` helper
**File**: `cmd/rpi/serve.go`
**Changes**: Add a helper function that takes a `*cobra.Command` and returns `Long` + `Example` combined:

```go
func mcpDescription(cmd *cobra.Command) string {
    desc := cmd.Long
    if cmd.Example != "" {
        desc += "\n\nExamples:\n" + cmd.Example
    }
    return desc
}
```

For multi-tool commands (frontmatter, git-context, verify, extract) where one Cobra command maps to multiple MCP tools, add a variant that prepends an action-specific prefix:

```go
func mcpDescriptionWithPrefix(prefix string, cmd *cobra.Command) string {
    return prefix + "\n\n" + mcpDescription(cmd)
}
```

#### 2. Replace all MCP tool descriptions
**File**: `cmd/rpi/serve.go`
**Changes**: Replace every `Description: "..."` in `registerTools()` with the appropriate `mcpDescription()` call.

**1:1 mappings** (10 tools):
| MCP Tool | Cobra Var |
|----------|-----------|
| `rpi_scan` | `scanCmd` |
| `rpi_scaffold` | `scaffoldCmd` |
| `rpi_chain` | `chainCmd` |
| `rpi_index_build` | `indexBuildCmd` |
| `rpi_index_query` | `indexQueryCmd` |
| `rpi_index_files` | `indexFilesCmd` |
| `rpi_index_status` | `indexStatusCmd` |
| `rpi_archive_scan` | `archiveScanCmd` |
| `rpi_archive_check_refs` | `archiveCheckRefsCmd` |
| `rpi_archive_move` | `archiveMoveCmd` |

**Multi-tool from parent** (10 tools):
| MCP Tool | Cobra Parent | Prefix |
|----------|-------------|--------|
| `rpi_git_context` | `gitContextCmd` | "Gather full git context." |
| `rpi_git_changed_files` | `gitContextCmd` | "List files changed vs main branch." |
| `rpi_git_sensitive_check` | `gitContextCmd` | "Scan staged files for sensitive content." |
| `rpi_frontmatter_get` | `frontmatterCmd` | "Read frontmatter fields from an artifact file." |
| `rpi_frontmatter_set` | `frontmatterCmd` | "Set a frontmatter field value." |
| `rpi_frontmatter_transition` | `frontmatterCmd` | "Validated status transition (enforces state machine)." |
| `rpi_verify_completeness` | `verifyCmd` | "Check plan progress: checkbox counts and file coverage." |
| `rpi_verify_markers` | `verifyCmd` | "Scan for TODO/FIXME/HACK markers in source files." |
| `rpi_extract` | `extractCmd` | "Extract a section from a markdown file." |
| `rpi_extract_list_sections` | `extractCmd` | "List all section headings in a markdown file." |

#### 3. Tests
**File**: `cmd/rpi/serve_test.go`
**Changes**: Add tests within the existing integration test or as new test functions:

- **GG-1 test**: Assert `rpi_scaffold` description contains "Types and their subdirectories" (from scaffold's `Long`)
- **GG-2 test**: Assert `rpi_chain` description contains "rpi chain .rpi/plans/" (from chain's `Example`)
- **GG-3 test**: Structural check — iterate all registered tools and verify none have a description shorter than 50 chars (one-liners would be short; Cobra Long fields are always long)
- **GG-4 test**: Assert `rpi_frontmatter_transition` description contains "draft" and "active" and "complete" (state transitions from parent `Long`)

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestMCPDescription` — new tests pass
- [x] `go test ./...` — all existing tests still pass

#### Manual Verification:
- [ ] Spot-check 2-3 MCP tool descriptions are rich (not one-liners)
- [ ] Verify no hardcoded description strings remain in `registerTools()`

### Commit:
- [x] Stage: `cmd/rpi/serve.go`, `cmd/rpi/serve_test.go`
- [x] Message: `feat(mcp): derive tool descriptions from Cobra Long+Example fields`

**Note**: Pause for manual confirmation before proceeding to next phase.

---

## Phase 2: Rewrite Command Prompts

### Overview
Rewrite all 7 command prompts to the goal/invariants/principles structure. Each prompt ≤50 lines (excluding YAML frontmatter), no `rpi_*` tool name references, no `rpi` CLI invocations. Preserves all workflow invariants, pipeline transitions, and mode detection. Addresses spec behaviors GG-5, GG-6, GG-7, GG-8, GG-9, GG-10.

### Tasks:

#### 1. Rewrite `rpi-research.md`
**File**: `internal/workflow/assets/commands/rpi-research.md`
**Changes**: Replace ~109 lines with ~30-line goal+guardrails version. Preserve:
- Conversational-first approach (no forced artifacts)
- Discovery interview before investigation
- Scale effort to question complexity
- Next stage → `/rpi-propose`

#### 2. Rewrite `rpi-propose.md`
**File**: `internal/workflow/assets/commands/rpi-propose.md`
**Changes**: Replace ~140 lines with ~40-50 line version. Preserve:
- Three modes: focused decision (quick), complex feature (full), updating existing (incremental)
- Mode auto-detection from input
- Design + spec as dual deliverables
- Artifact linking and transitions
- Next stage → `/rpi-plan`

#### 3. Rewrite `rpi-plan.md`
**File**: `internal/workflow/assets/commands/rpi-plan.md`
**Changes**: Replace ~114 lines with ~40-50 line version. Preserve:
- Two modes: standalone (simple tasks) vs pipeline (from design)
- Phased structure with success criteria
- Spec behavior mapping
- Upstream artifact transitions
- Next stage → `/rpi-implement`

#### 4. Rewrite `rpi-implement.md`
**File**: `internal/workflow/assets/commands/rpi-implement.md`
**Changes**: Replace ~162 lines with ~40-50 line version. Preserve:
- Plan validation and resumption (check existing checkmarks)
- Pre-review before each phase
- Red/green TDD for new code
- Sensitive file check before commits
- Spec conformance at completion
- Phase-by-phase commits with pause between phases

#### 5. Rewrite `rpi-verify.md`
**File**: `internal/workflow/assets/commands/rpi-verify.md`
**Changes**: Replace ~83 lines with ~30-40 line version. Preserve:
- Auto-detection from git changes when no path given
- Three dimensions: completeness, correctness, coherence
- Severity classification (blocker/warning/note)
- Spec conformance checking
- Purely advisory (never blocks)

#### 6. Rewrite `rpi-archive.md`
**File**: `internal/workflow/assets/commands/rpi-archive.md`
**Changes**: Replace ~154 lines with ~30-40 line version. Preserve:
- Two modes: specific paths vs scan
- Double confirmation for draft/active artifacts
- Cross-reference warnings
- Never auto-archive, never delete

#### 7. Rewrite `rpi-commit.md`
**File**: `internal/workflow/assets/commands/rpi-commit.md`
**Changes**: Replace ~72 lines with ~25-35 line version. Preserve:
- Sensitive file scanning before staging
- Logical commit grouping
- User approval before executing
- Hook failure → new commit (never amend)
- HEREDOC for commit messages

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` — all existing tests still pass

#### Manual Verification:
- [ ] Each prompt has exactly 3 sections: Goal, Invariants, Principles
- [ ] Each prompt is ≤50 lines (excluding frontmatter)
- [ ] No prompt contains `rpi_` or `` `rpi `` patterns
- [ ] Pipeline transitions preserved (research→propose→plan→implement)
- [ ] Mode detection preserved in propose and plan

### Commit:
- [x] Stage: all 7 `internal/workflow/assets/commands/rpi-*.md` files
- [x] Message: `refactor(prompts): rewrite commands to goal+guardrails structure`

**Note**: Pause for manual confirmation before proceeding to next phase.

---

## Phase 3: Add Prompt Structure Tests

### Overview
Add automated tests that enforce the spec's structural requirements on prompts. These serve as regression guards. Addresses all spec behaviors as automated checks.

### Tasks:

#### 1. Add prompt structure tests
**File**: `internal/workflow/workflow_test.go` (or existing test location)
**Changes**: Add test functions that load all 7 command prompt files and verify:

- **GG-5**: Each prompt contains `## Goal`, `## Invariants`, `## Principles` headings
- **GG-6**: Line count ≤50 (excluding YAML frontmatter block)
- **GG-7**: No matches for `rpi_` or backtick-quoted `rpi ` patterns

```go
func TestPromptStructure(t *testing.T) {
    // Load all rpi-*.md files from embedded assets
    // For each: check headings, line count, no tool references
}
```

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` — all tests pass including new prompt structure tests

#### Manual Verification:
- [ ] New tests correctly catch violations (temporarily break a prompt to verify)

### Commit:
- [x] Stage: test file(s)
- [x] Message: `test(prompts): add structural validation for goal+guardrails format`

---

## References
- Design: .rpi/designs/2026-03-21-goal-guardrails-prompt-rewrite.md
- Spec: .rpi/specs/goal-guardrails-prompt-rewrite.md
