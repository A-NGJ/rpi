---
domain: context-preservation
feature: context-preservation
last_updated: 2026-04-07T12:23:09+02:00
updated_by: .rpi/designs/2026-04-07-context-preservation-via-rpi-context-essentials.md
---

# Context Preservation

## Purpose

Provide a compact snapshot of the active implementation context — plan phase, spec scenarios, design constraints, and git state — so that AI assistants can recover after conversation compaction or context loss. Exposed as both an MCP tool and CLI command, with automatic triggering via Claude Code's PostCompact hook.

## Scenarios

### Auto-detect active plan when no path given
Given an active plan exists in the `.rpi/plans/` directory
When a user or tool calls `rpi_context_essentials` without a plan path
Then it returns context for the most recently dated active plan including its current phase and progress

### Explicit plan path returns that plan's context
Given a plan file exists at a known path
When a user or tool calls `rpi_context_essentials` with that plan path
Then it returns context specifically for that plan regardless of other active plans

### Current phase reflects first phase with unchecked items
Given a plan has phases where the first two phases are fully checked and the third has unchecked items
When context is retrieved for that plan
Then the current phase is reported as the third phase and the next unchecked items are listed

### Linked spec scenarios are included as titles
Given a plan links to a spec via frontmatter
When context is retrieved for that plan
Then the response includes the spec's feature name and a list of scenario titles without the full Given/When/Then text

### Design constraints are included
Given a plan links to a design that has a Constraints section
When context is retrieved for that plan
Then the response includes the constraints text from the linked design

### Empty result when no active plan exists
Given no plans have active status in the `.rpi/plans/` directory
When a user or tool calls `rpi_context_essentials` without a plan path
Then it returns an empty result indicating no active implementation context was found

### PostCompact hook reminds assistant to restore context
Given a project is initialized with Claude Code hooks via `rpi init`
When a conversation compaction occurs
Then the assistant receives guidance to call `rpi_context_essentials` to restore its implementation context

### CLI command outputs the same data as MCP tool
Given an active plan exists
When a user runs `rpi context` from the command line
Then the output matches the same JSON structure returned by the MCP tool

## Constraints
- Output must stay under approximately 500 tokens to fit alongside compacted conversation context
- All logic must live in the Go binary — hooks are thin triggers only
- Must reuse existing internals (scanner, chain resolver, checkbox parser, scenario parser) rather than duplicating logic
- Auto-detect selects the most recently dated active plan when multiple active plans exist
- Next unchecked items are capped at 3 entries to control output size
- Constraints text from the design is truncated if it exceeds 200 characters

## Out of Scope
- Session resume across all artifact types (dedicated `rpi_session_resume` tool — separate design)
- Pipeline step suggestion (`rpi_suggest_next` — separate feature)
- Plan boundary enforcement via hooks (`rpi_validate_action` — separate feature)
- Full Given/When/Then text in the context snapshot (too verbose for context injection)
