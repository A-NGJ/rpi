---
domain: rpi-status
feature: rpi-status
last_updated: 2026-04-04T22:30:00+02:00
updated_by: .rpi/designs/2026-04-04-rpi-status-dashboard-command.md
---

# rpi status dashboard command

## Purpose

Provide a single CLI command (`rpi status`) that aggregates all RPI artifact metadata into a human-readable dashboard showing artifact counts, stale warnings, active plan progress, and archive readiness.

## Scenarios

### Artifacts grouped by type and status
Given the `.rpi/` directory contains plans, designs, and research artifacts with various statuses
When the user runs `rpi status`
Then output shows non-archived artifacts grouped by type and status with counts, excluding specs and reviews from the artifacts summary

### Stale artifacts flagged by age
Given artifacts with draft or active status whose frontmatter date exceeds the staleness threshold
When the user runs `rpi status`
Then stale artifacts are listed with their path, status, and age in days, and the threshold is configurable via `--stale-days N`

### Active plans show checkbox progress
Given active or draft plans with checkbox items
When the user runs `rpi status`
Then each plan shows a checked/total count with percentage, and plans with zero checkboxes show no progress indicator

### Specifications section lists specs with scenario counts
Given specs exist in `.rpi/specs/`
When the user runs `rpi status`
Then a Specifications section lists each spec by name with its scenario count

### Archive-ready artifacts identified
Given artifacts with complete or superseded status and zero active references
When the user runs `rpi status`
Then they appear in a Ready to Archive section as a summary count grouped by type

### JSON output provides full structured data
Given any `.rpi/` state
When the user runs `rpi status --format json`
Then output is valid JSON with keys for summary, stale, active plans, and archivable data

### Empty sections are omitted
Given a `.rpi/` directory where some categories have no artifacts
When the user runs `rpi status`
Then sections with no data are not displayed at all

## Constraints
- Compose only existing internal packages
- Produce stable, deterministic output ordering (sorted by type, then path)
- Do not write to any files or modify artifact state
- Do not include archived artifacts in any section
- Do not emit color codes or Unicode special characters

## Out of Scope
- MCP tool registration
- Filtering flags (use `rpi scan` for filtered queries)
- Multi-level chain resolution (use `rpi chain` for full depth)
- File modification time — frontmatter dates only
