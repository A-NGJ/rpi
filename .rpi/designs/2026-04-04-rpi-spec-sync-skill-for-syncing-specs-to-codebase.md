---
date: 2026-04-04T22:50:56+02:00
status: complete
tags:
    - design
topic: rpi-spec-sync skill for syncing specs to codebase
---

# Design: rpi-spec-sync Skill

## Summary

A Claude skill (`/rpi-spec-sync`) that treats the codebase as ground truth and syncs specs to match it — the reverse of `/rpi-verify`. Uses existing MCP tools for structural analysis, then applies LLM judgment to rewrite, rename, delete, or keep each spec.

## Context

Specs drift from reality over time: features get removed, behaviors change, files get renamed. Currently there's no structured way to detect and fix this drift. `/rpi-verify` checks code against specs (specs are truth), but nothing checks specs against code (code is truth).

We recently migrated all specs to Gherkin-inspired scenario format and consolidated filenames. That process was manual — reading each spec, comparing to code, deciding what to do. This skill automates that workflow.

### Existing tools that support this

The infrastructure already exists — this skill orchestrates it:
- `rpi_verify_spec` — parse scenarios from a spec into structured JSON
- `rpi_scan` — discover all specs in `.rpi/specs/`
- `rpi_archive_check_refs` — find which artifacts reference a spec
- `rpi_git_changed_files` — detect recent code changes
- `rpi_chain` — resolve artifact dependency chains

## Constraints

- Skill only — no new Go code, CLI commands, or MCP tools
- Must use existing MCP tools for structural analysis
- Must get user confirmation before executing destructive actions (delete, rewrite)
- Must preserve spec history by archiving rather than deleting

## Components

### 1. Skill File

A new `rpi-spec-sync/SKILL.md` added to both `internal/workflow/assets/skills/` (embedded source) and installed to `.claude/skills/` via `rpi init`/`rpi update`.

The skill operates in two phases:

**Phase 1: Scan and assess** (structural, using MCP tools)
- Scan all specs via `rpi_scan --type spec`
- For each spec, parse scenarios via `rpi_verify_spec`
- Check for drift signals:
  - **Staleness**: spec `last_updated` significantly older than recent git activity on related code
  - **Dead references**: scenarios mention code paths, commands, or behaviors that no longer exist
  - **Orphaned specs**: zero incoming references from plans/designs (nothing links to this spec)
  - **Naming mismatch**: filename doesn't match the `feature` frontmatter field
- Present a drift report grouped by severity

**Phase 2: Act** (semantic, using LLM judgment + user confirmation)
For each flagged spec, read the actual implementation code and determine the action:
- **Keep**: spec is still accurate, no changes needed
- **Rewrite**: update scenarios to match current code behavior (preserve feature/domain, rewrite Given/When/Then)
- **Rename**: filename doesn't match feature field — rename file and update references
- **Merge**: two or more specs cover overlapping behavior — combine into a single spec, archive the sources
- **Archive**: feature was removed or spec is fully obsolete — archive via `rpi_archive_move`

Present proposed actions as a summary table, get user approval, then execute.

### 2. Drift Detection Heuristics

The skill uses these signals to flag specs for review (no single signal is conclusive — they're inputs to LLM judgment):

| Signal | How detected | What it suggests |
|--------|-------------|-----------------|
| Stale date | `last_updated` > 30 days behind git log on related files | Spec may not reflect recent changes |
| Dead scenario | Scenario references a command/feature that doesn't exist | Spec is outdated or obsolete |
| Orphaned | Zero references from active plans/designs | Spec may be abandoned |
| Naming mismatch | Filename ≠ `feature` field value | Should be renamed for consistency |
| Overlapping specs | Multiple specs reference the same domain or code area | Candidates for merge |

**Alternative considered**: Hard-coded rules (e.g., "auto-delete if orphaned + stale > 90 days") — rejected because deletion decisions need human judgment. The skill flags and suggests, user confirms.

## File Structure

| File | Action | Description |
|------|--------|-------------|
| `internal/workflow/assets/skills/rpi-spec-sync/SKILL.md` | Create | Embedded skill source |
| `.claude/skills/rpi-spec-sync/SKILL.md` | Create | Installed skill copy |

## Risks

**Over-flagging**: If heuristics are too sensitive, every spec gets flagged and the user has to review all of them. Mitigation: skill presents a summary first and asks which specs to investigate deeper, rather than deep-diving all at once.

**Rewrite quality**: LLM-generated scenario rewrites may lose important behavioral nuance from the original spec. Mitigation: skill shows the diff (old scenarios vs. new) for each rewrite and requires explicit approval.

## Out of Scope

- New CLI commands or MCP tools (skill-only)
- Automated/unattended execution (always requires user confirmation)
- Spec creation from scratch (that's `/rpi-propose`)
- Syncing non-spec artifacts (plans, designs, research)

## References

- Existing skill pattern: `.claude/skills/rpi-verify/SKILL.md`
- Existing skill pattern: `.claude/skills/rpi-archive/SKILL.md`
- Spec format spec: `.rpi/specs/spec-format.md`
