---
domain: session-awareness
feature: session-awareness
last_updated: 2026-04-07T15:00:00+02:00
updated_by: .rpi/designs/2026-04-07-session-awareness-tools.md
---

# Session Awareness

## Purpose

Give AI assistants awareness of active work and pipeline flow via two MCP tools: one that summarizes all active artifacts with implementation context on session start, and one that recommends the next pipeline step based on artifact state. Exposed as both MCP tools and CLI commands, with automatic triggering via Claude Code hooks.

## Scenarios

### Session resume returns active artifacts with status and context
Given active and draft artifacts exist across plans, designs, and research directories
When session resume is called
Then it returns a list of non-archived active and draft artifacts with their type, status, topic, and path

### Session resume includes active plan progress
Given an active plan exists with some phases fully checked and others with unchecked items
When session resume is called
Then it includes the current phase name, checked and total counts, and up to three next unchecked items for the most recently dated active plan

### Session resume includes a suggested next action
Given active artifacts exist in the `.rpi/` directory
When session resume is called
Then it includes a suggested action with reasoning derived from the pipeline suggestion engine

### Empty state when no active work exists
Given no active or draft artifacts exist in the `.rpi/` directory
When session resume is called
Then it returns an empty artifacts list, no plan context, and a suggestion to start new work

### Implementation continuation suggested for active plans
Given an active plan has unchecked items remaining
When the next pipeline action is requested
Then it suggests continuing implementation with the plan path and reports the number of remaining items

### Verification suggested when plan is fully checked
Given an active plan has all checkbox items checked and links to a spec
When the next pipeline action is requested
Then it suggests verifying the implementation against the linked spec

### Next pipeline step suggested for completed upstream artifacts
Given an artifact has reached its terminal status but no downstream artifact exists in the pipeline
When the next pipeline action is requested
Then it suggests creating the next downstream artifact referencing the completed one

## Constraints
- All logic lives in the Go binary — hooks are thin prompt-injection triggers only
- Reuses existing internals (scanner, chain resolver, checkbox parser, frontmatter parser) rather than duplicating logic
- Pipeline rules are evaluated in priority order: later pipeline stages take precedence over earlier stages
- When multiple artifacts compete at the same priority level, the most recently dated artifact wins
- Downstream artifact detection checks type-specific frontmatter fields (plan's `design` field, design's `related_research` field) rather than general reference counting
- Session resume includes plan context only for the most recently dated active plan
- Next unchecked items in plan context are capped at three entries

## Out of Scope
- Plan boundary enforcement during implementation (dedicated `rpi_validate_action` — separate feature)
- Spec drift detection (future mode on `rpi_verify_spec`)
- Writing plan progress (checking or unchecking items in plan files)
- Hook configuration for non-Claude Code tools (those use prompt guidance to call the MCP tools directly)
- Suggesting actions for archived or superseded artifacts
