---
archived_date: "2026-04-02"
branch: feature/goal+guardrails-prompt
date: 2026-03-21T00:39:48+01:00
git_commit: b3096e4
related_research: .thoughts/research/2026-03-13-showboat-like-pattern-for-rpi.md
repository: ai-agent-research-plan-implement-flow
researcher: Claude
status: archived
tags:
    - research
    - cli
    - ux
    - self-discovery
    - showboat
    - mcp
topic: showboat-pattern-for-rpi-prompt-rewrite
---

# Research: Showboat Pattern for RPI Prompt Rewrite

## Research Question
Can the RPI workflow commands adopt the showboat "goal + tool discovery" pattern — replacing rigid step-by-step prompt templates with short goal/guardrails prompts where the AI discovers how to use `rpi` tools on its own?

## Problem Statement
RPI command prompts (`.claude/commands/rpi-*.md`) are 100-140 line procedural scripts that spell out every tool invocation step-by-step. This makes them brittle, hard to maintain, and prevents the AI from adapting its approach to the situation. The showboat binary proves a different pattern works: a 3-line prompt that says "learn the tool via `--help`" lets the agent self-discover and adapt. The question is whether RPI's MCP tools expose enough information for this pattern to succeed.

## Summary

The showboat pattern works because its `--help` is 4.3KB of self-documenting text with examples, sample output, and edge cases. RPI already completed Phase 1-2 of help enrichment (all 18 Cobra commands have rich `Long` + `Example` fields), but the **MCP tool descriptions remain one-liners**. This is the critical gap — the agent sees `"Generate artifact files from templates"` instead of the rich help text that would enable self-discovery.

Two things are needed: (1) enrich MCP tool descriptions to match CLI help quality, (2) rewrite command prompts from procedural scripts to goal+guardrails format.

## Detailed Findings

### 1. Showboat's Pattern
- `/walkthrough` command is ~3 lines: "run `showboat --help` to learn showboat"
- Agent reads 4.3KB help, discovers commands, builds documents without further instruction
- Works because: verbose help with examples, simple append-only model, no wrong moves

### 2. Current RPI Prompt Structure
| Command | Lines | Explicit tool references |
|---------|-------|------------------------|
| rpi-propose | ~140 | 15 |
| rpi-plan | ~110 | 7 |
| rpi-implement | ~100 | 8 |
| rpi-verify | ~100 | 5 |
| rpi-research | ~110 | 5 |
| rpi-archive | ~80 | 4 |
| rpi-commit | ~40 | 2 |

### 3. MCP Tool Coverage
20 MCP tools cover all operations the current prompts need:
- **Artifact CRUD**: scaffold, scan, extract, extract_list_sections
- **Linking & tracing**: chain, frontmatter_get, frontmatter_set, frontmatter_transition
- **Verification**: verify_completeness, verify_markers
- **Git**: git_context, git_changed_files, git_sensitive_check
- **Index**: index_query, index_build, index_files, index_status
- **Archive**: archive_scan, archive_check_refs, archive_move

**Nothing is missing functionally.** Every operation in the rigid prompts maps to an available MCP tool.

### 4. The Discoverability Gap
MCP tool descriptions are one-liners:
- `rpi_scaffold` → "Generate artifact files from templates"
- `rpi_chain` → "Resolve artifact cross-reference chain"
- `rpi_frontmatter_transition` → "Validated status transition (enforces state machine)"

Missing from descriptions: valid parameter values, output format, workflow context, examples, edge cases. The CLI has all this in Cobra `Long` + `Example` fields — it just doesn't flow through to MCP.

### 5. Prior Work
- Research: `.thoughts/research/2026-03-13-showboat-like-pattern-for-rpi.md` — identified the pattern
- Proposal: `.thoughts/proposals/2026-03-13-self-documenting-help-for-agent-discovery.md` — designed CLI help enrichment
- Plan: `.thoughts/archive/2026-03/plan/2026-03-13-self-documenting-help-for-agent-discovery.md` — implemented (all 3 phases complete)
- Verification: `.thoughts/reviews/2026-03-15-verify-self-documenting-help-for-agent-discovery.md` — pass with 2 minor warnings

The CLI help is rich. The MCP descriptions are not. The prompt rewrite was identified as future work but never started.

### 6. Showboat vs RPI: Key Difference
Showboat is a document builder (append-only, no wrong moves). RPI orchestrates a multi-stage workflow with real constraints:
- State machine transitions (draft→active→complete→archived)
- Artifact linking (designs reference research, plans reference designs)
- Duplicate checking (scan before scaffold)
- Traceability (chain resolution)

This means pure "figure it out" won't work — the prompt needs **invariants** (must-do items) alongside the freedom.

## Assessment

### Benefits
1. Prompts drop from ~140 to ~20-40 lines — easier to maintain
2. Agent adapts approach to situation instead of following rigid script
3. Single source of truth — rich MCP descriptions serve both discovery and documentation
4. New tools become automatically discoverable without prompt rewrites
5. Better agent judgment on investigation depth, checkpointing, parallelization

### Risks and Mitigations
| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Agent skips critical invariants (linking, dedup, transitions) | Medium | High | Explicit invariants section in each prompt (5-8 must-do items) |
| Agent picks wrong tool/params without guidance | Low | Medium | Rich MCP descriptions with valid values and examples |
| Different models perform differently with loose prompts | Medium | Medium | Incremental rollout — rewrite one command first, test across models |
| Regression detection harder | Low | Low | Keep original prompts in version control as fallback |
| Agent over/under-explores | Medium | Low | Principles section guides effort calibration |

### Fail-safes already in place
- `rpi_frontmatter_transition` rejects invalid transitions
- `rpi_scaffold` auto-populates frontmatter with dates, git info
- `rpi_chain` detects cycles and has depth limits

## Suggested Next Steps

1. **Propose** a two-part change:
   - Part A: Enrich MCP tool descriptions (port Cobra `Long` + `Example` content to MCP server descriptions)
   - Part B: Rewrite command prompts to goal+guardrails format
2. Part A should land first — it's prerequisite for Part B
3. Start Part B with `/rpi-research` (simplest command) as proof of concept

## Decisions

- The "guardrails not rails" approach is the right middle ground between rigid templates and pure discovery
- MCP description enrichment is the prerequisite — without it, the agent can't self-discover
- Incremental rollout (one command at a time) mitigates risk
