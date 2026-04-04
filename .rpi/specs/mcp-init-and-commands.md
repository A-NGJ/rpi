---
domain: mcp-init-and-commands
feature: mcp-init
last_updated: 2026-04-04T22:30:00+02:00
updated_by: .rpi/archive/2026-03/design/2026-03-18-mcp-native-commands-and-init.md
---

# MCP Init and Commands

## Purpose

Ensure `rpi init` configures the MCP server and all embedded workflow files reference MCP tool names instead of CLI commands.

## Scenarios

### Init registers MCP server when prerequisites are met
Given `rpi` and `claude` are both found in PATH and no MCP server is registered
When the user runs `rpi init`
Then the MCP server is registered via `claude mcp add rpi -- rpi serve`

### Init handles missing CLI gracefully
Given `rpi` or `claude` is not resolvable in PATH
When the user runs `rpi init`
Then MCP configuration is skipped with a warning and init completes successfully

### Init skips MCP with --no-mcp flag
Given any system state
When the user runs `rpi init --no-mcp`
Then MCP registration is skipped entirely regardless of PATH availability

### Init does not overwrite existing MCP entry
Given the `rpi` MCP server is already registered
When the user runs `rpi init`
Then the existing entry is preserved and a warning is printed

### Embedded files use MCP tool references not CLI commands
Given all embedded command and skill files in the workflow assets
When scanning their content
Then no file contains imperative instructions to shell out to `rpi` CLI commands or PATH prerequisites

## Constraints
- MCP registration uses `claude mcp add` to register the server
- Do not write MCP config when `--target opencode` is used
- Do not overwrite an already-registered MCP server
- Do not modify the MCP server implementation

## Out of Scope
- OpenCode MCP configuration
- New MCP tools
- PIPELINE.md template updates
