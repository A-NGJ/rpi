---
archived_date: "2026-04-02"
date: 2026-03-21T23:53:36+01:00
spec: .rpi/specs/rpi-explain.md
status: archived
tags:
    - design
topic: rpi-explain command for post-implementation walkthroughs
---

# Design: rpi-explain command for post-implementation walkthroughs

## Summary

A new `/rpi-explain` slash command that generates a diff-scoped walkthrough of an implemented solution, highlighting non-obvious changes with explanations. It fills the gap between `/rpi-verify` (checks correctness) and `/walkthrough` (explains all source code) by focusing specifically on what changed and why.

## Context

After `/rpi-implement` completes, a developer may want to deeply understand the solution — either because they're reviewing someone else's work, revisiting their own after time, or simply want a sanity check on the "why" behind non-obvious decisions.

Current tools don't cover this:
- **`/walkthrough`** — walks through entire source files, not scoped to a diff
- **`/rpi-verify`** — checks spec conformance with severity ratings, doesn't explain intent
- **Design/plan artifacts** — capture planned rationale, but implementations often diverge

The explain command bridges this by reading the actual diff, cross-referencing plan/design artifacts, and producing an explanation focused on non-obvious parts.

## Constraints

- Must work without upstream artifacts (plan/design) — the diff alone is sufficient input
- Must not duplicate `/rpi-verify` — no severity ratings, no pass/fail judgments
- Should be lightweight — a prompt-only command (`.claude/commands/`), no new RPI CLI tooling needed
- Artifact saving is optional and user-triggered, not automatic

## Components

### 1. `/rpi-explain` slash command (`.claude/commands/rpi-explain.md`)

The command prompt that instructs Claude to:

1. **Resolve context**: Accept an optional artifact path or auto-detect from recent git changes. If an artifact chain exists (plan → design → research), read it for context.
2. **Get the diff**: Use `rpi git-context changed-files` to identify changed files, then read the actual diff.
3. **Generate walkthrough**: Walk through the diff file-by-file, explaining:
   - What changed (factual summary)
   - Why it changed (inferred from artifacts or code context)
   - Non-obvious decisions flagged with explicit callouts
4. **Optional artifact save**: If asked, save to `.rpi/reviews/` as an explanation artifact.

No new RPI CLI commands are needed. The existing tooling is sufficient:
- `rpi git-context changed-files` — identifies what changed
- `rpi chain <artifact>` — resolves upstream context
- `rpi scaffold verify-report` — could scaffold an explanation artifact (reusing review type)

**Alternative considered**: A new `rpi explain` CLI command that gathers the diff + artifact context in one call. Rejected because the value is in the LLM's explanation, not in deterministic tooling — a prompt-only command keeps it simple.

### 2. No new RPI CLI tooling required

The existing MCP tools cover all data-gathering needs:
- `rpi_git_changed_files` — file list
- `rpi_chain` — artifact chain resolution
- `rpi_scaffold` — artifact creation (verify-report type for optional saving)

The only new artifact is the command file itself.

## File Structure

| File | Action | Purpose |
|------|--------|---------|
| `.claude/commands/rpi-explain.md` | Create | Slash command prompt |

## Risks

- **Low signal-to-noise**: If the diff is large, the explanation could become verbose. Mitigation: the command should prioritize non-obvious changes and summarize straightforward ones briefly.
- **Hallucinated rationale**: Without upstream artifacts, the LLM might invent reasons for changes. Mitigation: command instructions should distinguish "inferred from artifacts" vs "inferred from code context" and flag uncertainty.

## Out of Scope

- New RPI CLI commands or Go code
- New artifact types (reuse `verify-report` scaffold if saving)
- Automated triggering after `/rpi-implement` — this is always user-initiated
- Comparison against specific branches/commits (uses default `changed-files` behavior)

## References

- Existing commands: `.claude/commands/rpi-verify.md`, `.claude/commands/rpi-implement.md`
- RPI CLI: `cmd/rpi/gitcontext.go`, `cmd/rpi/chain.go`
