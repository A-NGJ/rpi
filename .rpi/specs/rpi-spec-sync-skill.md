---
domain: rpi-spec-sync
feature: spec-sync
last_updated: 2026-04-04T22:50:56+02:00
updated_by: .rpi/designs/2026-04-04-rpi-spec-sync-skill-for-syncing-specs-to-codebase.md
---

# rpi-spec-sync Skill

## Purpose

A Claude skill that syncs specs to match the current codebase — the reverse of verification. Uses existing MCP tools for structural drift detection, then applies LLM judgment to rewrite, rename, archive, or keep each spec.

## Scenarios

### Scan identifies stale specs
Given specs exist in `.rpi/specs/` and some have not been updated in over 30 days while related code has changed
When the user runs `/rpi-spec-sync`
Then a drift report is presented showing which specs are flagged with the reason for each flag

### Scan identifies obsolete specs
Given a spec describes a feature that has been removed from the codebase
When the user runs `/rpi-spec-sync`
Then the spec is flagged for archival with an explanation of what's missing

### Scan identifies naming mismatches
Given a spec's filename does not match its `feature` frontmatter field
When the user runs `/rpi-spec-sync`
Then the spec is flagged for rename with the suggested new filename

### User approves actions before execution
Given the drift report contains flagged specs with proposed actions (rewrite, rename, archive, keep)
When the user reviews the proposals
Then no changes are made until the user explicitly approves each action or approves all at once

### Rewrite updates scenarios to match code
Given a spec is approved for rewrite
When the skill executes the rewrite
Then the scenarios are updated to reflect current code behavior while preserving the spec's domain, feature name, and constraints

### Archive removes obsolete specs cleanly
Given a spec is approved for archival
When the skill executes the archive
Then the spec is moved to `.rpi/archive/` and all references in other artifacts are updated

### Rename updates filename and references
Given a spec is approved for rename
When the skill executes the rename
Then the file is renamed to match the feature field and all references in other artifacts are updated

### Merge combines overlapping specs
Given two or more specs cover closely related or overlapping behavior
When the skill flags them for merge and the user approves
Then the scenarios are combined into a single spec with a unified feature name, the source specs are archived, and all references are updated

## Constraints
- Must use existing MCP tools for structural analysis — no new CLI commands or Go code
- Must get user confirmation before any destructive action (rewrite, archive, rename)
- Rewrites preserve the spec's domain, feature field, and constraints — only scenarios change
- Archives use the existing archive flow (move to `.rpi/archive/`, update frontmatter)
- Reference updates cover all `.rpi/` artifacts (plans, designs, reviews, archives)

## Out of Scope
- Creating new specs from scratch (that's `/rpi-propose`)
- Syncing non-spec artifacts (plans, designs, research)
- Unattended/automated execution without user confirmation
