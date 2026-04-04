---
date: 2026-04-04T01:42:56+02:00
status: active
tags:
    - design
topic: 'unified status lifecycle: remove approved/implemented, add reopen'
---

# Design: Unified Status Lifecycle

## Summary

Two changes: (1) Collapse the two parallel status pipelines into one universal lifecycle with `complete → active` for re-opening. (2) Specs become statusless living documents — only archivable on explicit demand.

## Context

The state machine currently has two pipelines:
- **Plans/Designs**: `draft → active → complete → archived`
- **Specs**: `draft → approved → implemented → archived`

In practice, specs don't follow their dedicated pipeline. `index-expansion.md` has `status: complete` (not `implemented`), and 6 other implemented specs remain `active` — the `approved`/`implemented` statuses aren't used consistently because the distinction doesn't carry its weight.

Specs are living documents. They don't "complete" — they describe current behavior and evolve as features change. Giving them a status lifecycle creates false signals (is it `active`? `complete`? does it matter?). The right model: specs exist, and get archived only when explicitly retired.

## Constraints

- The state machine remains global for artifact types that use it
- `superseded` must remain a universal escape hatch
- `archived` must remain terminal
- Specs must still be archivable via `rpi archive move` on demand

## Components

### State Machine (`internal/frontmatter/transition.go`)

**Before:**
```
draft       → active, approved, superseded
active      → complete, superseded
approved    → implemented, superseded
complete    → archived, superseded
implemented → archived, superseded
```

**After:**
```
draft    → active, superseded
active   → complete, superseded
complete → active, archived, superseded
```

Removes `approved` and `implemented` entirely. Adds `complete → active` (reopen).

### Specs: No Status Field

Specs become statusless living documents:
- Spec scaffold templates (both `.rpi/templates/spec.tmpl` and `internal/workflow/assets/templates/spec.tmpl`) drop the `status` field from frontmatter
- Existing specs with `status` fields: the field becomes inert (not read or acted upon by any code path)

### Archivable Filter (`internal/scanner/scan.go:123-128`)

**Before:** Specs only archivable when `superseded`; others when `complete`/`superseded`/`implemented`
**After:** Specs are never surfaced by the archivable scanner. Other types archivable when `complete` or `superseded`. `implemented` removed.

Specs can still be archived manually via `rpi archive move`.

### Status Display (`cmd/rpi/status.go:63`)

Remove `implemented` from `statusDisplayOrder`.

### MCP Tool Description (`cmd/rpi/serve.go:313`)

Update the `status` field description to list only valid statuses: `draft, active, complete, superseded, archived`.

### Skills (`.claude/skills/`)

| Skill | Change |
|-------|--------|
| `rpi-propose` | Remove spec status transition entirely (specs have no status) |
| `rpi-propose` | "approved spec" → "spec" |
| `rpi-implement` | Remove spec status transition entirely |
| `rpi-implement` | "approved plan" → "active plan" |

### Test Fixtures

| File | Change |
|------|--------|
| `internal/frontmatter/frontmatter_test.go` | Rewrite valid/invalid transition cases — remove `approved`/`implemented`, add `complete → active` |
| `internal/scanner/scan_test.go:37` | Change spec fixture to have no status (or remove status field) |
| `internal/scanner/scan_test.go:152` | Update archivable test — specs are never archivable via scanner |
| `internal/chain/resolve_test.go:277` | Change `status: approved` to no status or `active` |

## Risks

- **Existing specs with status fields**: Inert but present. Could be cleaned up in a follow-up or left alone — the field is simply ignored.
- **Re-opening abuse for non-specs**: `complete → active` could be misused. Mitigated by `superseded` existing for the "replaced, don't touch" case.

## Out of Scope

- Automated transition triggers (all transitions remain manual)
- Migration tooling for `approved`/`implemented` (no live artifacts use them)
- Bulk cleanup of existing spec `status` fields (inert, harmless)
