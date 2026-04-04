---
domain: unified status lifecycle
feature: status-lifecycle
last_updated: 2026-04-04T22:30:00+02:00
updated_by: .rpi/designs/2026-04-04-unified-status-lifecycle-remove-approved-implemented-add-reopen.md
---

# Unified Status Lifecycle

## Purpose

Simplify the status lifecycle: one universal pipeline for plans/designs/research with reopen support, and specs and reviews as statusless documents archivable only on demand.

## Scenarios

### Valid transitions between statuses succeed
Given an artifact at any non-terminal status
When transitioning to an allowed target (draft→active, draft→superseded, active→complete, active→superseded, complete→active, complete→archived, complete→superseded)
Then the transition succeeds and the status is updated

### Legacy statuses are rejected
Given an artifact at any status
When attempting to transition to `approved` or `implemented`
Then the transition fails with a validation error

### Archived is terminal
Given an artifact with status `archived`
When attempting to transition to any other status
Then the transition fails with a validation error

### Complete can be reopened to active
Given an artifact with status `complete`
When transitioning to `active`
Then the transition succeeds, enabling rework on previously completed artifacts

### Specs and reviews excluded from archivable scanner
Given specs or reviews in `.rpi/` with or without status fields
When scanning for archivable artifacts
Then specs and reviews never appear in the results regardless of their fields

### Specs and reviews can be archived manually
Given a spec or review in `.rpi/`
When the user runs `rpi archive move` on it
Then the artifact is moved to the archive directory regardless of status

## Constraints
- State machine is global — no per-type branching
- `superseded` is reachable from draft, active, and complete
- Missing status is treated as `draft`
- Do not allow transitions out of `archived`
- Do not surface specs or reviews in archivable scanner results

## Out of Scope
- Automated transition triggers
- Bulk cleanup of existing spec status fields
- Per-type state machines
