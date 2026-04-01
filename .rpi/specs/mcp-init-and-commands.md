---
domain: mcp-init-and-commands
id: MC
last_updated: 2026-03-18T00:00:00Z
status: active
updated_by: .rpi/proposals/2026-03-18-mcp-native-commands-and-init.md
---

# MCP Init and Commands

## Purpose

Ensure `rpi init` configures the MCP server and all embedded workflow files (commands, skills) reference MCP tool names instead of CLI commands.

## Behavior

### Init MCP Configuration
- **MC-1**: `rpi init` registers the MCP server via `claude mcp add rpi -- rpi serve` when target is `claude` and both `rpi` and `claude` are found in PATH
- **MC-2**: If `rpi` or `claude` is not in PATH, `rpi init` warns and skips MCP configuration without failing
- **MC-3**: If the `rpi` MCP server is already registered, `rpi init` warns and does not overwrite the existing entry
- **MC-5**: `rpi init --no-mcp` skips MCP configuration entirely

### Commands MCP References
- **MC-7**: No embedded command file in `internal/workflow/assets/commands/` contains the text "must be available in PATH"
- **MC-8**: No embedded command file contains instructions to shell out to `rpi` (no "Run `rpi ", "run `rpi ", or backtick-quoted `rpi <subcommand>` patterns as imperative instructions)
- **MC-9**: Each command file that previously referenced `rpi` operations now references the corresponding `rpi_*` MCP tool name

### Skills MCP References
- **MC-10**: No embedded skill file in `internal/workflow/assets/skills/` contains imperative instructions to run `rpi` CLI commands
- **MC-11**: Skill files reference MCP tool names (`rpi_index_query`, `rpi_scan`, etc.) where they previously referenced CLI commands

### Init Flag
- **MC-12**: `rpi init` accepts a `--no-mcp` boolean flag (default false)

## Constraints

### Must
- MCP registration uses `claude mcp add` to register the server
- All 7 command files and 4 skill files are updated in the embedded assets (source of truth)
- `rpi init --update` does NOT re-configure MCP (only rebuilds index and CLI reference)

### Must Not
- Write MCP config when `--target opencode` is used
- Overwrite an already-registered `rpi` MCP server
- Break any existing `rpi init` tests
- Modify the MCP server implementation (`serve.go`)

### Out of Scope
- OpenCode MCP configuration
- New MCP tools
- PIPELINE.md template updates (cosmetic, not behavioral)

## Test Cases

### MC-1: Init registers MCP server
- **Given** `rpi` and `claude` are in PATH and no `rpi` MCP server is registered **When** `rpi init` runs **Then** the MCP server is registered via `claude mcp add rpi -- rpi serve`

### MC-2: Init skips MCP when rpi or claude not in PATH
- **Given** `rpi` or `claude` is not resolvable via `exec.LookPath` **When** `rpi init` runs **Then** MCP config step is skipped, a warning is printed, and init completes successfully

### MC-3: Init warns on existing rpi MCP entry
- **Given** the `rpi` MCP server is already registered **When** `rpi init` runs **Then** the existing entry is preserved and a warning is printed

### MC-5: Init skips MCP with --no-mcp
- **Given** any state **When** `rpi init --no-mcp` runs **Then** MCP registration is skipped entirely

### MC-7: Commands have no PATH prerequisite
- **Given** all embedded command files **When** scanning for "must be available in PATH" **Then** zero matches found

### MC-8: Commands have no shell-out instructions
- **Given** all embedded command files **When** scanning for imperative `rpi` CLI invocation patterns **Then** zero matches found

### MC-9: Commands reference MCP tools
- **Given** the `rpi-propose.md` command **When** reading its content **Then** it contains `rpi_scaffold`, `rpi_frontmatter_transition`, `rpi_scan`, `rpi_chain` tool references

### MC-12: --no-mcp flag exists
- **Given** `rpi init` command definition **When** checking registered flags **Then** `--no-mcp` flag is present with default value `false`
