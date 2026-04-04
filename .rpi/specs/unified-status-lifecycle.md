---
domain: unified status lifecycle
id: UL
last_updated: 2026-04-04T01:43:32+02:00
updated_by: .rpi/designs/2026-04-04-unified-status-lifecycle-remove-approved-implemented-add-reopen.md
---

# Unified Status Lifecycle

## Purpose

Simplify the status lifecycle: one universal pipeline for plans/designs/research with reopen support, and specs and reviews as statusless documents archivable only on demand.

## Behavior

### State Machine
- **UL-1**: Valid statuses are: `draft`, `active`, `complete`, `superseded`, `archived`
- **UL-2**: `approved` and `implemented` are not valid statuses — transitions to or from them are rejected
- **UL-3**: Allowed transitions: `draft → active`, `draft → superseded`, `active → complete`, `active → superseded`, `complete → active`, `complete → archived`, `complete → superseded`
- **UL-4**: `archived` is terminal — no transitions out
- **UL-5**: Missing status is treated as `draft`
- **UL-6**: `complete → active` (reopen) is valid for all artifact types that have status

### Specs and Reviews Are Statusless
- **UL-7**: Spec and review scaffold templates do not include a `status` field in frontmatter
- **UL-8**: Specs and reviews are never surfaced by the archivable scanner, regardless of any fields present
- **UL-9**: Specs and reviews can still be archived manually via `rpi archive move`
- **UL-10**: Skills do not transition spec or review status

### Archivable Filter
- **UL-11**: Non-spec, non-review artifacts are archivable when status is `complete` or `superseded`
- **UL-12**: `implemented` is not a valid archivable status

### Status Display
- **UL-13**: `rpi status` display order includes only: `active`, `draft`, `complete`, `superseded`

### MCP Tool Description
- **UL-14**: The `rpi_frontmatter_transition` tool description lists only valid target statuses: `draft`, `active`, `complete`, `superseded`, `archived`

## Constraints

### Must
- State machine remains global for types that use it — no per-type branching
- All existing valid transitions (except those involving `approved`/`implemented`) continue to work
- `superseded` remains reachable from `draft`, `active`, and `complete`
- `rpi archive move` works on specs and reviews regardless of status/absence of status

### Must Not
- Must not allow transitions out of `archived`
- Must not accept `approved` or `implemented` as valid transition targets
- Must not surface specs or reviews in archivable scanner results

### Out of Scope
- Automated transition triggers
- Bulk cleanup of existing spec `status` fields
- Per-type state machines

## Test Cases

### UL-2: Rejected legacy statuses
- **Given** an artifact with status `draft` **When** transitioning to `approved` **Then** transition fails with ValidationError
- **Given** an artifact with status `active` **When** transitioning to `implemented` **Then** transition fails with ValidationError

### UL-3: All allowed transitions
- **Given** an artifact at each status **When** transitioning to each allowed target **Then** all transitions in UL-3 succeed and all others fail

### UL-4: Archived is terminal
- **Given** an artifact with status `archived` **When** transitioning to any status **Then** transition fails with ValidationError

### UL-6: Reopen from complete
- **Given** an artifact with status `complete` **When** transitioning to `active` **Then** transition succeeds and status is `active`

### UL-8: Specs and reviews excluded from archivable scan
- **Given** a spec with no status field and zero references **When** scanning with archivable filter **Then** the spec does not appear in results
- **Given** a spec with `status: complete` (legacy) and zero references **When** scanning with archivable filter **Then** the spec does not appear in results
- **Given** a review with no status field and zero references **When** scanning with archivable filter **Then** the review does not appear in results

### UL-11: Non-spec archivable
- **Given** a design with status `complete` and zero references **When** scanning with archivable filter **Then** the design appears in results
