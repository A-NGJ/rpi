---
date: 2026-03-21T23:18:23+01:00
topic: "goal-guardrails-prompt-rewrite"
tags: [verify]
design: .rpi/designs/2026-03-21-goal-guardrails-prompt-rewrite.md
plan: .rpi/plans/2026-03-21-goal-guardrails-prompt-rewrite.md
spec: .rpi/specs/goal-guardrails-prompt-rewrite.md
---

# Verification Report: goal-guardrails-prompt-rewrite

## Summary

**Status: PASS** — Implementation fully matches the design and spec across all three dimensions. All 10 spec behaviors (GG-1 through GG-10) are satisfied. Full test suite passes. No blockers.

- Blockers: 0
- Warnings: 1
- Notes: 2

## Completeness

### Plan Progress
- **Checkboxes**: 10/18 checked (8 unchecked are "Manual Verification" success criteria — verified below)
- **File coverage**: All 10 plan files match git changed files exactly. No missing, no unexpected.
- **TODO/FIXME/HACK markers**: 3 found, all false positives (the word "TODO" appears in description text for the marker-scanning tool itself, not as actual work items)

### Phase 1: MCP Description Enrichment — COMPLETE
- `mcpDescription()` and `mcpDescriptionWithPrefix()` implemented at `cmd/rpi/serve.go:70-83`
- All 20 tool registrations use these functions — zero hardcoded description strings in `registerTools()`
- Tests GG-1 through GG-4 pass (`TestMCPDescription_*` in `cmd/rpi/serve_test.go:449-583`)
- Spot-checked: `rpi_scaffold`, `rpi_chain`, `rpi_git_context`, `rpi_frontmatter_set` all have rich multi-line descriptions with examples

### Phase 2: Rewrite Command Prompts — COMPLETE
All 7 prompts rewritten. Verification of unchecked success criteria:

| Criterion | Result |
|-----------|--------|
| Each prompt has Goal/Invariants/Principles | PASS — all 7 have exactly these 3 sections |
| Each prompt ≤50 lines (excl. frontmatter) | PASS — range: 24 (commit) to 35 (propose, plan) |
| No `rpi_` or `` `rpi `` patterns | PASS — zero matches across all 7 files |
| Pipeline transitions preserved | PASS — research→propose, propose→plan, plan→implement all present |
| Mode detection preserved | PASS — propose has 3 modes (focused/complex/incremental), plan has 2 modes (standalone/pipeline) |

### Phase 3: Prompt Structure Tests — COMPLETE
- `TestPromptStructure_HasRequiredSections` (GG-5) at `internal/workflow/workflow_test.go:183`
- `TestPromptStructure_LineCount` (GG-6) at `internal/workflow/workflow_test.go:199`
- `TestPromptStructure_NoToolReferences` (GG-7) at `internal/workflow/workflow_test.go:224`
- All pass in `go test ./...`

## Correctness

### Spec Behavior Verification

| Behavior | Status | Evidence |
|----------|--------|----------|
| GG-1: MCP descriptions include Long text | PASS | `TestMCPDescription_IncludesLongText` — checks `rpi_scaffold` contains "Types and their subdirectories" |
| GG-2: MCP descriptions include Example text | PASS | `TestMCPDescription_IncludesExamples` — checks `rpi_chain` contains "rpi chain .rpi/plans/" |
| GG-3: Single source of truth | PASS | `TestMCPDescription_SingleSourceOfTruth` — all tools have descriptions >50 chars; code inspection confirms zero inline strings |
| GG-4: Frontmatter tools include parent context | PASS | `TestMCPDescription_FrontmatterIncludesParent` — checks `rpi_frontmatter_transition` contains "draft", "active", "complete" |
| GG-5: Prompt structure (3 sections) | PASS | `TestPromptStructure_HasRequiredSections` + manual verification |
| GG-6: Prompt length ≤50 lines | PASS | `TestPromptStructure_LineCount` + manual verification (max: 35 lines) |
| GG-7: No tool name references | PASS | `TestPromptStructure_NoToolReferences` + manual grep |
| GG-8: Invariants preserved | PASS | Manual review: propose's Invariants section covers all original requirements (existing design check, research chain, codebase investigation, buy-in, artifact linking, spec creation, transitions, next stage) |
| GG-9: Pipeline transitions | PASS | research→`/rpi-propose`, propose→`/rpi-plan`, plan→`/rpi-implement` |
| GG-10: Mode detection | PASS | propose: focused/complex/incremental; plan: standalone/pipeline |

### API Contracts
- No CLI behavioral changes — only description text and prompt content changed
- `mcpDescription()` signature: `func mcpDescription(cmd *cobra.Command) string` — straightforward, no side effects
- `mcpDescriptionWithPrefix()` signature: `func mcpDescriptionWithPrefix(prefix string, cmd *cobra.Command) string` — prefix + Long + Example

## Coherence

- **Naming**: `mcpDescription`/`mcpDescriptionWithPrefix` follow Go conventions and the existing codebase style
- **Code organization**: Helper functions placed near `registerTools()` in `serve.go` — logical grouping
- **Error handling**: No new error paths introduced; description generation is pure string concatenation
- **Test patterns**: New tests follow the same `newRPIServer()` + in-memory transport pattern used by `TestIntegration_AllToolsRegistered`
- **Prompt style**: All 7 prompts follow a consistent structure with uniform formatting conventions (bold labels, arrow transitions, em-dash separators)
- **No unnecessary dependencies**: Pure refactoring — no new imports or dependencies added

## Issues

### Blockers

None.

### Warnings

1. **Plan checkboxes not fully updated** — 8 "Manual Verification" checkboxes remain unchecked in the plan file despite the work being complete. This is cosmetic but makes the plan appear incomplete when scanned by tools.
   - File: `.rpi/plans/2026-03-21-goal-guardrails-prompt-rewrite.md:102,188-192,230`

### Notes

1. **TODO markers are false positives** — The 3 TODO markers found by `rpi verify markers` are in description text for the marker-scanning tool itself (`cmd/rpi/serve.go:166,531`) and the verify command prompt (`rpi-verify.md:20`). They describe the tool's purpose, not actual work items.

2. **Out-of-scope template changes detected in git status** — `.claude/templates/plan.tmpl` and `.claude/templates/verify-report.tmpl` appear in `git status` as modified but are not part of this plan's scope. These may be unrelated changes.

## References
- Design: .rpi/designs/2026-03-21-goal-guardrails-prompt-rewrite.md
- Plan: .rpi/plans/2026-03-21-goal-guardrails-prompt-rewrite.md
- Spec: .rpi/specs/goal-guardrails-prompt-rewrite.md
- Research: .rpi/research/2026-03-21-showboat-pattern-for-rpi-prompt-rewrite.md
