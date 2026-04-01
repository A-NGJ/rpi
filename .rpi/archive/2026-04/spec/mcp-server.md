---
archived_date: "2026-04-02"
domain: mcp-server
id: MS
last_updated: 2026-03-17T23:47:30+01:00
status: archived
updated_by: .rpi/proposals/2026-03-17-mcp-server.md
---

# MCP Server

## Purpose

Expose all RPI CLI operations as typed MCP tools over stdio, enabling AI assistants to call them with validated schemas instead of constructing free-form shell commands.

## Behavior

### Server Lifecycle
- **MS-1**: `rpi serve` starts an MCP server on stdio transport and blocks until the client disconnects
- **MS-2**: The server advertises implementation name `"rpi"` and a version matching the binary version

### Tool Registration
- **MS-3**: The server registers exactly one MCP tool per CLI leaf command (20 tools total)
- **MS-4**: Each tool name uses the `rpi_` prefix with underscores replacing spaces and hyphens (e.g., `rpi_frontmatter_get`)
- **MS-5**: Each tool has a non-empty description matching its CLI `Short` help text
- **MS-6**: Input schemas are auto-generated from Go structs with `json` and `jsonschema` tags

### Tool Execution
- **MS-7**: Each tool handler calls the same internal package function as the corresponding CLI command
- **MS-8**: All tools return results as JSON-formatted `TextContent`
- **MS-9**: Tool errors are returned as MCP errors (not as JSON with error fields in TextContent)

### Tool Parity
- **MS-10**: `rpi_scan` accepts optional `type`, `status`, `proposal`, `references`, `archivable` fields and returns the same JSON as `rpi scan --format json`
- **MS-11**: `rpi_scaffold` accepts `type` (required), `topic` (required), and optional `ticket`, `research`, `proposal`, `spec`, `tags`, `write`, `force` fields. Returns the rendered artifact content or the written file path
- **MS-12**: `rpi_frontmatter_get` accepts `file` (required) and optional `field`. Returns all frontmatter as JSON, or a single field value
- **MS-13**: `rpi_frontmatter_set` accepts `file`, `field`, `value` (all required). Writes the value and returns success
- **MS-14**: `rpi_frontmatter_transition` accepts `file` and `status` (both required). Enforces the same state machine as the CLI
- **MS-15**: `rpi_chain` accepts `path` (required) and optional `sections`. Returns the same chain JSON as the CLI
- **MS-16**: `rpi_extract` accepts `path` and `section` (both required). Returns the section content as JSON
- **MS-17**: `rpi_extract_list_sections` accepts `path` (required). Returns section headings as JSON
- **MS-18**: `rpi_verify_completeness` accepts `plan_path` (required). Returns checkbox + file coverage JSON
- **MS-19**: `rpi_verify_markers` accepts optional `file_path`. Returns marker scan JSON
- **MS-20**: `rpi_index_build` accepts optional `lang`, `path`, `force`. Builds the index and returns summary
- **MS-21**: `rpi_index_query` accepts `pattern` (required) and optional `kind`, `exported`. Returns matching symbols as JSON
- **MS-22**: `rpi_index_files` accepts optional `lang`. Returns indexed files as JSON
- **MS-23**: `rpi_index_status` accepts no parameters. Returns index metadata as JSON
- **MS-24**: `rpi_git_context` accepts no parameters. Returns full git context JSON
- **MS-25**: `rpi_git_changed_files` accepts no parameters. Returns changed file list as JSON
- **MS-26**: `rpi_git_sensitive_check` accepts no parameters. Returns sensitive file scan results as JSON
- **MS-27**: `rpi_archive_scan` accepts no parameters. Returns archivable artifacts with ref counts as JSON
- **MS-28**: `rpi_archive_check_refs` accepts `path` (required). Returns referencing files as JSON
- **MS-29**: `rpi_archive_move` accepts `path` (required) and optional `force`. Returns move result JSON

## Constraints

### Must
- All tools must return identical data to the equivalent CLI command with `--format json`
- The server must not modify any global state on startup (only tool calls have side effects)
- The existing CLI must continue to work unchanged

### Must Not
- Must not expose a `format` parameter on any MCP tool (always JSON)
- Must not bundle multiple CLI commands into a single MCP tool with an `action` discriminator

### Out of Scope
- SSE or HTTP transport
- Authentication
- Skill file portability across AI clients (separate effort)
- CLI deprecation

## Test Cases

### MS-1: Server starts on stdio
- **Given** `rpi serve` is invoked **When** the process starts **Then** it blocks listening on stdin/stdout until client disconnect

### MS-3: All tools registered
- **Given** a running MCP server **When** the client sends `tools/list` **Then** the response contains exactly 20 tools, all prefixed with `rpi_`

### MS-7: Tool calls internal package
- **Given** a running MCP server and a `.rpi/` directory with artifacts **When** the client calls `rpi_scan` with `{"type": "proposal"}` **Then** the response matches `rpi scan --type proposal --format json`

### MS-9: Errors returned as MCP errors
- **Given** a running MCP server **When** `rpi_frontmatter_get` is called with a non-existent file **Then** the response is an MCP error, not a JSON body with an error field

### MS-11: Scaffold returns file path when write=true
- **Given** a running MCP server **When** `rpi_scaffold` is called with `{"type": "research", "topic": "test", "write": true}` **Then** the response contains the path to the created file

### MS-14: Transition enforces state machine
- **Given** a running MCP server and an artifact with status `draft` **When** `rpi_frontmatter_transition` is called with status `complete` **Then** the tool returns an MCP error indicating invalid transition
