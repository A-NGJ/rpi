---
date: 2026-04-08T12:14:26+02:00
related_research: .rpi/research/2026-04-08-documentation-restructure-for-external-users.md
spec: .rpi/specs/documentation-structure.md
status: complete
tags:
    - design
topic: documentation restructure for external users
---

# Design: documentation restructure for external users

## Summary

Restructure README.md as a landing page that hooks external developers browsing GitHub. Reorder sections to lead with value, add a hero example, badges, and a concrete Quick Start ending. No changes to docs/ files.

## Context

Research ([.rpi/research/2026-04-08-documentation-restructure-for-external-users.md]) analyzed 6 popular dev tool READMEs and found that RPI's documentation content is solid but packaged as internal docs. The README reads like a reference for existing users, not a landing page for new ones. The best content (concrete workflow examples) is buried in `docs/workflow-guide.md`.

The core problem: a developer scrolling past on GitHub can't tell in 5 seconds why they'd want this.

## Constraints

- README must remain a single file (no splitting into multiple READMEs)
- docs/ files stay as-is — they work well as supporting references
- No visual assets (GIF/screencast) in this design — separate task requiring external tooling
- Must work for both Claude Code and OpenCode users (current dual-target support preserved)

## Components

### 1. Pain-focused opener

**Current** (README.md:1-3): Mechanism-focused — "A structured development workflow for AI coding agents that turns vague feature requests into shipped code through a pipeline of discrete, reviewable stages."

**New**: Lead with the control problem, then the solution:

> AI coding agents are capable — the challenge is steering them. Without structure, you end up re-running prompts, hoping the output lands closer to what you need. RPI gives you a framework to direct that capability: staged decisions, reviewable artifacts, and specs that keep the work on track.

This frames RPI as empowering (not compensating for bad AI), focuses on the user's control problem, and names the concrete mechanisms.

### 2. Badges

Add 3 badges immediately after the H1, before the opener text:

- **License**: MIT badge (static, from shields.io)
- **Release**: Latest release version (dynamic, from GitHub API via shields.io)
- **CI**: Release workflow status badge (from `.github/workflows/release.yml`)

These are standard social proof signals that every popular tool uses. Quick win, zero maintenance.

### 3. Hero example

Pull from `docs/workflow-guide.md:36-68` (the rate limiting medium-task example) and condense into a 4-step sequence showing the full flow:

```
1. Research (optional)  →  /rpi-research How does the API middleware chain work?
2. Propose              →  /rpi-propose Add per-endpoint rate limiting
3. Plan                 →  /rpi-plan .rpi/designs/2026-03-04-api-rate-limiting.md
4. Implement            →  /rpi-implement .rpi/plans/2026-03-04-api-rate-limiting.md
```

Each step gets: the command, one line describing what happens, and what artifact is produced. The point is to show the full cycle in under 20 lines — a reader should understand the workflow pattern from this alone.

This replaces the ASCII artifact-flow diagram (README.md:7-14), which communicates structure but not experience.

### 4. "How RPI is different" — repositioned comparison

Move from the bottom (currently line 116) to right after the hero example. Rename from "How It Compares" to "How RPI Is Different" — confident positioning rather than defensive comparison.

Content stays largely the same but the framing shifts: lead with the two differentiators (reviewable artifacts + compiled CLI), then the vs. comparisons.

### 5. Concrete Quick Start ending

Current "Start coding" step (README.md:69-71) dead-ends with "use the slash commands." Replace with a concrete first command and expected result:

```
### 3. Try it

/rpi-plan Fix the date formatter in utils/dates.ts that returns "NaN" for ISO strings

Claude investigates the file, writes a plan to `.rpi/plans/`, then:

/rpi-implement .rpi/plans/2026-04-08-fix-date-formatter.md

Review the changes, approve, done.
```

This mirrors the small-task path from `docs/workflow-guide.md:11-29` — the lowest-friction entry point.

### 6. Section reorder

**Current order**: Title → Opener → ASCII diagram → Why → Quick Start → Slash Commands → Choosing Your Path → Documentation → MCP Server → How It Compares → Acknowledgments → License

**New order**: Title → Badges → Opener → Hero Example → How RPI Is Different → Quick Start (with concrete ending) → Slash Commands → Choosing Your Path → Documentation → MCP Server → Acknowledgments → License

Key changes:
- Badges and opener move above the fold
- Hero example replaces ASCII diagram (moved to just after opener)
- Comparison moves from bottom to after hero example
- "Why This Exists" is removed as a standalone section — its best points are absorbed into the opener and the "How RPI Is Different" section
- Quick Start gets concrete ending

## File Structure

| File | Change |
|------|--------|
| README.md | Restructured — new section order, opener, hero example, badges, concrete Quick Start |
| docs/* | No changes |

## Risks

- **Hero example gets stale**: The rate limiting example references specific command patterns. If slash command names change, the example breaks. Mitigation: the example uses stable command names that are unlikely to change, and workflow-guide.md will catch the same drift.
- **Badges break on repo rename/transfer**: shields.io badges reference the GitHub owner/repo. Mitigation: standard risk for any project, easily fixable.

## Out of Scope

- Visual assets (GIF, screencast, terminal recording) — requires separate tooling (asciinema/vhs), tracked as a future task
- docs/ file restructuring — research found these are fine as-is
- Content additions to docs/ files (new guides, tutorials)
- Website or documentation site (e.g., MkDocs, Docusaurus)

## References
- Research: .rpi/research/2026-04-08-documentation-restructure-for-external-users.md
