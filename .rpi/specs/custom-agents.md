---
domain: custom-agents
feature: custom-agents
last_updated: 2026-04-29T00:00:00+02:00
updated_by: .rpi/plans/2026-04-29-remove-worktree-mode-from-rpi-implement.md
---

# Custom Agents

## Purpose

Provide a Claude Code custom agent — a read-only verification agent — so that spec conformance checks can run in parallel with implementation.

## Scenarios

### Init installs agent definitions for Claude target
Given a project being initialized with the Claude target
When `rpi init` completes
Then `.claude/agents/` contains `rpi-verify.md`

### Init skips agent definitions for non-Claude targets
Given a project being initialized with the agents-only or opencode target
When `rpi init` completes
Then no agent definition files are installed

### Update syncs agent definitions
Given a project already initialized for the Claude target
When `rpi update` runs
Then new or changed agent definitions are installed to `.claude/agents/` without overwriting user modifications unless forced

### Verification agent restricted to read-only operations
Given the installed verification agent definition
When inspecting its tool restrictions
Then it permits only read, search, and RPI MCP tools — no file creation or modification tools

### Verification agent returns structured results
Given an active plan with linked specs
When the verification agent is spawned with a plan path
Then it checks each spec scenario against actual code, reports pass/fail per scenario with file references, and returns an overall verdict

## Constraints

- Agent definitions are Claude Code-specific; they are never installed for non-Claude targets
- Verification agent must not modify any files — enforcement via allowed-tools restriction
- No new MCP tools — agents orchestrate existing tools only

## Out of Scope

- Plugin packaging (bundling agents + skills + hooks into one installable unit)
- Per-skill hooks for boundary enforcement
- Artifact navigator agent
- OpenCode or agents-only agent equivalents
- Verification report file generation (the agent returns inline results; the `/rpi-verify` skill handles report files)
