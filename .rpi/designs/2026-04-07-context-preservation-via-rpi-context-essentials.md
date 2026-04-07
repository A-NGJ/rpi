---
date: 2026-04-07T12:23:09+02:00
related_research: .rpi/research/2026-04-07-rpi-improvements-via-claude-code-internals.md
spec: .rpi/specs/context-preservation.md
status: complete
tags:
    - design
    - mcp
    - hooks
    - context
topic: context preservation via rpi-context-essentials
---

# Design: Context Preservation via rpi_context_essentials

## Summary

Add an `rpi_context_essentials` MCP tool (and `rpi context` CLI command) that returns a compact snapshot of the active implementation context — current plan phase, progress, linked spec scenario titles, design constraints, and git state. Wire it into Claude Code's `PostCompact` hook so context is automatically re-injected after conversation compaction. Non-Claude Code tools call the same MCP tool via prompt guidance.

## Context

Long `/rpi-implement` sessions frequently lose context when conversations compact. The AI assistant forgets which plan phase it's on, what the spec scenarios are, and what constraints apply. This is the most painful failure mode in RPI's current workflow — it causes repeated file reads, wrong-phase work, and spec drift.

The research (.rpi/research/2026-04-07-rpi-improvements-via-claude-code-internals.md) identified this as the highest-ROI improvement: Claude Code's `PostCompact` hook can re-inject essential state automatically, and the intelligence lives in an MCP tool that any client can call.

## Constraints

- **Token budget**: Output must stay under ~500 tokens when serialized — it gets injected as a system message alongside the compacted conversation
- **MCP-first**: All logic lives in the Go binary. The Claude Code hook is a thin trigger, not a logic layer
- **Portability**: The MCP tool works identically across Claude Code, Cursor, OpenCode, and any MCP client. Only the trigger mechanism differs
- **Reuse existing internals**: Build on `scanner.Scan`, `chain.Resolve`, `parseCheckboxes`, `parseScenarios`, and `git.GatherContext` — no duplicated logic
- **No new packages**: The context assembly logic fits in `cmd/rpi/context.go` alongside the existing command pattern

## Components

### 1. `rpi context` CLI Command + `rpi_context_essentials` MCP Tool

A single implementation exposed as both CLI command and MCP tool (same pattern as `rpi verify`).

**Input**: Optional `plan_path` string. When omitted, auto-detects by scanning for active plans via `scanner.Scan(rpiDir, Filters{Type: "plan", Status: "active"})`. If multiple active plans exist, returns context for the most recently modified one.

**Output structure**:

```json
{
  "plan": {
    "path": ".rpi/plans/2026-04-07-context-preservation.md",
    "topic": "context preservation via rpi-context-essentials",
    "current_phase": "Phase 2: MCP Tool Registration",
    "progress": { "checked": 5, "total": 12 },
    "next_items": [
      "Add contextInput struct with JSON schema tags",
      "Add handleContext handler"
    ]
  },
  "spec": {
    "path": ".rpi/specs/context-preservation.md",
    "feature": "context-preservation",
    "scenario_titles": [
      "Auto-detect active plan when no path given",
      "Explicit plan path returns that plan's context",
      "..."
    ]
  },
  "constraints": "- Token budget: under 500 tokens\n- MCP-first: logic in Go binary\n- ...",
  "git": {
    "branch": "feat/context-preservation",
    "uncommitted_files": 3
  }
}
```

**Auto-detection considered two alternatives:**

- **Scan-based (chosen)**: `scanner.Scan` with `type=plan, status=active`, pick most recent by `date` frontmatter. Simple, uses existing code, works when no plan path is available (e.g., after full context loss).
- **Git-based**: Infer from branch name or recent commits. Rejected — too fragile, branch naming is not standardized, and `rpi_git_context` already provides raw git info.

**"Current phase" detection**: Reuses `parseCheckboxes` from `verify.go`. Walk phases in order; the first phase containing unchecked items is the current phase. Extract up to 3 next unchecked items (configurable truncation to stay within token budget).

**Spec scenario titles**: Resolve chain from plan to find the linked spec, then use `parseScenarios` to extract titles only (not full Given/When/Then — too verbose for context injection).

**Design constraints extraction**: Resolve chain to find the linked design, extract `## Constraints` section via `frontmatter.ExtractSection`. Truncate to 200 characters if needed.

**Git state**: Lightweight — just branch name and count of uncommitted files. Full git context is available via `rpi_git_context` if needed.

### 2. PostCompact Hook Configuration

Added to `.claude/settings.json` by `rpi init` (Claude target only). The hook calls the `rpi_context_essentials` MCP tool after every conversation compaction.

**Hook configuration** (added to settings.json by `rpi init`):

```json
{
  "hooks": {
    "PostCompact": [
      {
        "type": "command",
        "command": "cat <<'HOOK_EOF'\nIMPORTANT: Context was compacted. Call the rpi_context_essentials MCP tool to restore your implementation context (active plan phase, spec scenarios, constraints).\nHOOK_EOF"
      }
    ]
  }
}
```

**Why a prompt injection instead of a direct MCP call**: Claude Code hooks can return text that gets appended to the conversation as `additionalContext`. A hook that echoes a reminder is simpler and more robust than one that tries to call an MCP tool directly — the AI assistant then calls the tool itself, which goes through normal MCP request/response flow.

**Alternative considered — direct tool call in hook**: The hook could use `curl` or `rpi context` to get the data and inject it directly. Rejected because: (a) the MCP server may not be running as a separate process (it's stdio-based), (b) calling `rpi context` via CLI in a hook adds latency and shell complexity, (c) the AI assistant calling the MCP tool is the intended portable pattern.

### 3. Prompt Guidance for Non-Claude Code Tools

Skill prompts (`rpi-implement/SKILL.md`) get an additional invariant:

> If context seems lost or you're unsure which phase you're on, call `rpi_context_essentials` to restore your implementation context.

This works for any MCP-capable tool. Non-MCP tools can use `rpi context` via CLI.

### 4. `rpi init` Integration

The `syncProject` function (called by both `rpi init` and `rpi update`) adds the PostCompact hook to `.claude/settings.json` when the target is Claude. Same merge pattern as the existing `configureSettings` function that adds `mcp__rpi__*` permissions.

## File Structure

| File | Change |
|---|---|
| `cmd/rpi/context.go` | **New** — CLI command, context assembly logic, output structs |
| `cmd/rpi/context_test.go` | **New** — tests for phase detection, auto-detect, truncation |
| `cmd/rpi/serve.go` | **Modified** — register `rpi_context_essentials` MCP tool |
| `cmd/rpi/init_cmd.go` | **Modified** — add PostCompact hook to settings.json merge |
| `.claude/skills/rpi-implement/SKILL.md` | **Modified** — add context recovery invariant |
| `internal/workflow/assets/skills/rpi-implement/SKILL.md` | **Modified** — same (embedded copy) |

## Risks

- **Multiple active plans**: If a user has several active plans, auto-detect picks the most recent. This could be wrong. Mitigation: the `plan_path` parameter lets the caller be explicit, and the hook reminder tells the assistant to call the tool (which returns the plan path), giving it a chance to correct.
- **Stale context after compaction**: The context snapshot reflects the plan file's current state. If the file wasn't updated (checkboxes not ticked), the "current phase" may be behind reality. Mitigation: `/rpi-implement` already updates checkboxes as items complete — this is an existing invariant.
- **Token budget creep**: If plans or specs grow large, the output could exceed 500 tokens. Mitigation: truncate `next_items` to 3, `scenario_titles` to titles only, `constraints` to 200 chars. These limits are hardcoded initially; can be made configurable later.

## Out of Scope

- **`rpi_session_resume`**: A dedicated tool for "what should I work on next?" across all artifact types. Separate design — `rpi_context_essentials` covers the active-implementation case well enough.
- **`rpi_suggest_next`**: Pipeline step suggestion based on artifact state. Separate feature.
- **`rpi_validate_action`**: Plan boundary enforcement via `PreToolUse` hook. Separate feature.
- **Plugin packaging**: Bundling all RPI config as a Claude Code plugin. Separate feature.
- **Custom agent definitions**: Verification or navigation agents. Separate feature.

## References
- Research: .rpi/research/2026-04-07-rpi-improvements-via-claude-code-internals.md
