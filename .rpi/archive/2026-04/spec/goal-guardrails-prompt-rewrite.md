---
archived_date: "2026-04-02"
domain: goal-guardrails-prompt-rewrite
id: GG
last_updated: 2026-03-21T00:44:38+01:00
status: archived
updated_by: .rpi/designs/2026-03-21-goal-guardrails-prompt-rewrite.md
---

# Goal+Guardrails Prompt Rewrite

## Purpose

Ensure MCP tool descriptions enable agent self-discovery and that rewritten command prompts preserve workflow correctness while being significantly shorter and more adaptive.

## Behavior

### MCP Description Enrichment
- **GG-1**: Every MCP tool description includes the corresponding Cobra command's `Long` field content
- **GG-2**: Every MCP tool description includes the corresponding Cobra command's `Example` field content (when present)
- **GG-3**: MCP descriptions are generated from Cobra commands at registration time — no hardcoded duplicate text
- **GG-4**: Frontmatter subcommand tools (get/set/transition) each include the parent command's `Long` field plus a one-line prefix identifying which action the tool performs

### Command Prompt Structure
- **GG-5**: Each rewritten command prompt contains exactly three sections: Goal, Invariants, Principles
- **GG-6**: Each rewritten command prompt is ≤50 lines (excluding YAML frontmatter)
- **GG-7**: No rewritten prompt contains explicit MCP tool names (e.g., `rpi_scaffold`, `rpi_scan`) or explicit `rpi` CLI invocations
- **GG-8**: Each prompt's Invariants section captures all must-do workflow steps from the original prompt (artifact linking, duplicate checking, status transitions, upstream validation)

### Workflow Correctness
- **GG-9**: The pipeline stage transitions are preserved: each prompt identifies its stage and suggests the correct next stage command with artifact path
- **GG-10**: Mode detection logic is preserved where applicable (propose: quick/full/incremental; plan: standalone/pipeline)

## Constraints

### Must
- MCP descriptions are derived from Cobra command objects — single source of truth
- All 7 command prompts are rewritten using the goal+guardrails structure
- Existing tests continue to pass (`go test ./...`)
- No CLI behavioral changes — only description text and prompt content change

### Must Not
- Prompts must not reference specific MCP tool names or `rpi` subcommand invocations
- MCP descriptions must not be hardcoded strings that duplicate Cobra `Long`/`Example` content
- Prompts must not exceed 50 lines (excluding frontmatter) — enforces conciseness

### Out of Scope
- New RPI tools (note/pop/exec for incremental building)
- Template (`.tmpl`) changes
- CLAUDE.md changes
- CLI behavioral changes

## Test Cases

### GG-1: MCP descriptions include Long text
- **Given** the MCP server is initialized **When** `rpi_scaffold` tool description is read **Then** it contains "Types and their subdirectories" (from scaffold's `Long` field)

### GG-2: MCP descriptions include Example text
- **Given** the MCP server is initialized **When** `rpi_chain` tool description is read **Then** it contains "rpi chain .rpi/plans/" (from chain's `Example` field)

### GG-3: Single source of truth
- **Given** `serve.go` **When** inspecting tool registration code **Then** every Description value is produced by calling `mcpDescription()` with a Cobra command — no inline description strings

### GG-4: Frontmatter tools include parent context
- **Given** the MCP server is initialized **When** `rpi_frontmatter_transition` description is read **Then** it contains the valid state transitions (draft→active, active→complete, etc.) from the parent command's `Long` field

### GG-5: Prompt structure
- **Given** each rewritten command prompt **When** parsed for `## ` headings **Then** it contains Goal, Invariants, and Principles sections

### GG-6: Prompt length
- **Given** each rewritten command prompt **When** lines are counted (excluding YAML frontmatter block) **Then** the count is ≤50

### GG-7: No tool name references
- **Given** each rewritten command prompt **When** searched for `rpi_` or `` `rpi `` patterns **Then** zero matches are found

### GG-8: Invariants preserved
- **Given** the original `rpi-propose.md` requires: check existing designs, read research chain, investigate codebase, get buy-in, link artifacts, create spec, transition statuses, suggest next stage **When** the rewritten version's Invariants section is read **Then** each of these requirements appears as an invariant item

### GG-9: Pipeline transitions preserved
- **Given** each rewritten prompt **When** the Goal section is read **Then** it identifies the current pipeline stage and names the next stage command to suggest

### GG-10: Mode detection preserved
- **Given** the rewritten `rpi-propose.md` **When** the Goal section is read **Then** it mentions the three modes (focused decision, complex feature, updating existing design) and how to detect them
