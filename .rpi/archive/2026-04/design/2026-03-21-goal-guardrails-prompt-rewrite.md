---
archived_date: "2026-04-02"
date: 2026-03-21T00:43:09+01:00
related_research: .rpi/research/2026-03-21-showboat-pattern-for-rpi-prompt-rewrite.md
status: archived
tags:
    - design
    - cli
    - ux
    - self-discovery
    - mcp
    - prompts
topic: goal-guardrails-prompt-rewrite
---

# Design: Goal+Guardrails Prompt Rewrite

## Summary

Replace rigid step-by-step RPI command prompts with short goal+guardrails prompts, and enrich MCP tool descriptions so the AI can self-discover how to use `rpi` tools. Two parts: (A) generate MCP descriptions from existing Cobra `Long`+`Example` fields — single source of truth, zero duplication; (B) rewrite 7 command prompts from ~100-140 lines to ~20-40 lines using a goal/invariants/principles structure.

## Context

RPI command prompts currently spend ~30-40% of their content on explicit tool invocation instructions ("Use the rpi_scaffold tool to scaffold..."). This makes them brittle, hard to maintain, and prevents the AI from adapting its approach. The showboat binary proves a different pattern: its `/walkthrough` command is 3 lines — "learn the tool via `--help`" — and the agent figures out the rest.

Prior work already enriched the CLI's Cobra help text (all 18 commands have rich `Long` + `Example` fields). But the MCP server still serves one-liner descriptions. The content exists; it just doesn't flow through to where the agent sees it.

## Constraints

1. **Cobra fields are the source of truth** — MCP descriptions must be generated from them, not duplicated
2. **Workflow invariants must be preserved** — artifact linking, duplicate checking, status transitions, traceability. These are the "guardrails"
3. **Tool fail-safes already exist** — `frontmatter_transition` rejects invalid transitions, `scaffold` auto-populates frontmatter. The tools enforce correctness even if the agent forgets a step
4. **Incremental rollout** — start with one command to validate the approach before rewriting all 7
5. **Backward compatible** — no CLI behavior changes, no MCP schema changes. Only description text and prompt content change

## Components

### Component A: MCP Description Enrichment

**What**: In `serve.go:registerTools()`, replace one-liner `Description` strings with content generated from the corresponding Cobra command's `Long` and `Example` fields.

**Alternative considered**: Copy-paste CLI help text into MCP descriptions manually. Rejected because it creates two copies that will drift.

**How it works**: Create a helper function that takes a `*cobra.Command` and returns a combined description string:

```go
func mcpDescription(cmd *cobra.Command) string {
    desc := cmd.Long
    if cmd.Example != "" {
        desc += "\n\nExamples:\n" + cmd.Example
    }
    return desc
}
```

Then in `registerTools()`:
```go
mcp.AddTool(s, &mcp.Tool{
    Name:        "rpi_scaffold",
    Description: mcpDescription(scaffoldCmd),
}, handleScaffold)
```

**Mapping**: Most tools map 1:1 to Cobra commands. The frontmatter subcommands (get/set/transition) are three MCP tools from one Cobra command — each gets the parent's `Long` (which documents all three actions) plus its specific action description.

### Component B: Goal+Guardrails Prompt Rewrite

**What**: Rewrite 7 command prompt files from procedural step-by-step scripts to a 3-section structure:

1. **Goal** (~3-5 lines) — What artifacts to produce, what pipeline stage this is, what to suggest next
2. **Invariants** (~5-8 lines) — Must-do items as a checklist. These are the non-negotiable steps the agent must not skip
3. **Principles** (~3-5 lines) — Quality guidance: be interactive, ground in evidence, right-size effort

**Example** — `/rpi-propose` rewritten:

```markdown
# Solution Design

Create a design document (`.rpi/designs/`) and behavioral spec (`.rpi/specs/`)
for the given topic or research. This is part of: research → **propose** → plan → implement.

## Invariants
- Check for existing designs on this topic before creating a new one
- If a research doc is provided, read it and resolve its artifact chain
- Investigate the codebase before proposing — ground decisions in evidence with file:line refs
- Get user buy-in on trade-offs before writing the design
- Link the design to upstream research via frontmatter
- Create a behavioral spec (SDD template) with test cases — present for approval
- Transition artifacts: design → active, research → complete (if fully addressed), spec → approved
- When done, suggest: → /rpi-plan <design-path>

## Principles
- Be interactive — checkpoint before major decisions
- Be opinionated — recommend with clear reasoning
- Scale effort to complexity — a focused decision needs less investigation than a new subsystem
- Specs are the contract — every design culminates in an approved spec
```

That's ~20 lines vs the current ~140. The agent discovers `rpi_scaffold`, `rpi_scan`, `rpi_chain`, `rpi_frontmatter_transition` from the enriched MCP tool descriptions.

**What stays in prompts**: Mode detection logic (quick/full/incremental for propose) stays as a brief note in the Goal section — it's workflow guidance, not tool instruction.

**What's removed**: All explicit tool references ("Use the rpi_scaffold tool to..."), numbered sub-steps within steps, repeated instructions that appear in multiple prompts.

### Rollout Strategy

1. Ship Component A (MCP enrichment) first — it's a prerequisite and zero-risk
2. Rewrite `/rpi-research` as proof of concept (simplest command, 5 tool references)
3. Validate with multiple models if possible
4. Rewrite remaining 6 commands one at a time
5. Keep original prompts in git history as fallback

## File Structure

| File | Change |
|------|--------|
| `cmd/rpi/serve.go` | Replace one-liner descriptions with `mcpDescription()` calls |
| `internal/workflow/assets/commands/rpi-research.md` | Rewrite to goal+guardrails (proof of concept) |
| `internal/workflow/assets/commands/rpi-propose.md` | Rewrite to goal+guardrails |
| `internal/workflow/assets/commands/rpi-plan.md` | Rewrite to goal+guardrails |
| `internal/workflow/assets/commands/rpi-implement.md` | Rewrite to goal+guardrails |
| `internal/workflow/assets/commands/rpi-verify.md` | Rewrite to goal+guardrails |
| `internal/workflow/assets/commands/rpi-archive.md` | Rewrite to goal+guardrails |
| `internal/workflow/assets/commands/rpi-commit.md` | Rewrite to goal+guardrails |

No new files created.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Agent skips invariants (linking, dedup, transitions) | High — broken artifact chains | Explicit invariants checklist in each prompt; tool-level fail-safes already reject invalid transitions |
| Different models perform differently with loose prompts | Medium — inconsistent quality | Incremental rollout; start with one command; keep originals as fallback |
| Frontmatter subcommands share one Cobra command but are 3 MCP tools | Low — description may be verbose | Include parent Long for all three; add a one-line prefix clarifying which action this tool performs |
| Agent over-explores without step-by-step guidance | Low — wastes time | Principles section guides effort calibration ("scale effort to complexity") |

## Out of Scope

- **Adding new RPI tools** (note/pop/exec for incremental building) — separate concern from the original showboat research
- **CLAUDE.md changes** — the command prompts are self-contained; CLAUDE.md doesn't need updating
- **Template changes** — artifact templates (`.tmpl` files) are unchanged
- **CLI behavior changes** — only description text changes, no functional changes

## References

- Research: .rpi/research/2026-03-21-showboat-pattern-for-rpi-prompt-rewrite.md
- Prior research: .thoughts/research/2026-03-13-showboat-like-pattern-for-rpi.md
- Prior proposal (help enrichment): .thoughts/proposals/2026-03-13-self-documenting-help-for-agent-discovery.md
- Prior implementation: .thoughts/archive/2026-03/plan/2026-03-13-self-documenting-help-for-agent-discovery.md
- MCP server: `cmd/rpi/serve.go`
- Cobra commands: `cmd/rpi/*.go`
