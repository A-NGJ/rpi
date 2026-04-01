---
date: 2026-03-17T23:46:25+01:00
status: complete
tags:
    - proposal
    - mcp
topic: mcp-server
---

# Proposal: MCP Server for RPI CLI

## Summary

Add an `rpi serve` subcommand that exposes all RPI CLI operations as MCP (Model Context Protocol) tools over stdio. This eliminates hallucinated subcommands/flags by giving AI assistants typed tool schemas instead of free-form shell commands, and makes the workflow portable across any MCP-compatible client (Claude Code, Cursor, Windsurf, Cline).

## Investigation Findings

- **The problem is structural**: AI assistants construct `rpi` shell commands from memory of `cli-reference.md` and sometimes hallucinate flags or subcommands. No amount of documentation prevents this — the model is composing free-form strings.
- **Clean internal separation**: `cmd/rpi/*.go` handles Cobra flag parsing + output formatting; `internal/*/` packages contain all business logic. The MCP server can call internal packages directly, making it an alternative frontend with no refactoring needed.
- **All commands already output JSON**: Every subcommand supports `--format json` (most default to it). MCP tools just drop the format parameter and always return JSON.
- **20 leaf commands exist**: `scan`, `scaffold`, `frontmatter get/set/transition`, `chain`, `extract`/`extract --list-sections`, `verify completeness/markers`, `index build/query/files/status`, `git-context`/`changed-files`/`sensitive-check`, `archive scan/check-refs/move`, `spec coverage`.
- **Alternatives considered and rejected**:
  - *PreToolUse hooks*: Claude Code-specific, not portable
  - *Better error messages*: Works everywhere but costs a wasted round-trip per mistake; corrective, not preventive
  - *Richer CLI reference docs*: Reduces hallucinations but can't eliminate them structurally

## Constraints & Requirements

- Must not break the existing CLI — `rpi serve` is additive
- Must work over stdio transport (how MCP clients launch tool servers)
- Must produce typed JSON schemas from Go structs so AI clients see exact parameter names, types, and descriptions
- Single binary — no separate `rpi-mcp` executable
- Minimal new dependencies

## Design Decisions

### Decision 1: Tool granularity — flat (one tool per CLI leaf command)

**Chosen**: Flat — 20 MCP tools, one per CLI leaf command (`rpi_scan`, `rpi_scaffold`, `rpi_frontmatter_get`, etc.)
**Alternatives considered**: Coarse grouping with `action` parameter (e.g., `rpi_frontmatter` with `action: "get"|"set"|"transition"`)
**Rationale**: The whole point is preventing hallucination. An `action` parameter reintroduces the "which action?" guessing problem. 20 tools is well within MCP's comfort zone, and each tool gets a self-contained, specific schema.

### Decision 2: Output format — always JSON text

**Chosen**: Always return JSON as `TextContent`
**Alternatives considered**: Structured output via the SDK's generic `Out` type parameter
**Rationale**: Consumers are AI models, not programs. They parse JSON from text naturally. The internal packages already produce JSON via `json.MarshalIndent`. Using the `Out` type adds schema complexity for no benefit.

### Decision 3: Server entry point — `rpi serve` subcommand

**Chosen**: `rpi serve` Cobra subcommand running MCP server over stdio
**Alternatives considered**: Separate `rpi-mcp` binary
**Rationale**: Single binary, discoverable via `rpi serve --help`, no separate build/install. MCP client config is `{"command": "rpi", "args": ["serve"]}`.

### Decision 4: Skill file updates — reference MCP tool names

**Chosen**: Update skill instructions to reference MCP tool names (e.g., "Use the `rpi_scaffold` tool") instead of shell commands
**Alternatives considered**: Keep as-is (Bash fallback), or remove all rpi references from skills
**Rationale**: Skills define the workflow — *when* to scaffold, *when* to transition. They should reference the tools by name. The skill format (`.claude/commands/*.md`) is still Claude Code-specific, so full portability is a separate effort. This is the right incremental step.

### Decision 5: Drop `--format` from MCP tools

**Chosen**: MCP tools always return JSON; no `format` parameter in input structs
**Alternatives considered**: Expose format as an optional parameter
**Rationale**: MCP tools are consumed by machines. JSON is the only useful format. Dropping the parameter simplifies schemas and removes a source of confusion.

## Architecture

```
AI Client (Claude Code, Cursor, etc.)
    │
    │  MCP protocol (JSON-RPC over stdio)
    │
    ▼
rpi serve (MCP Server)
    │
    │  Direct Go function calls
    │
    ▼
internal/* packages
    ├── scanner/     (scan)
    ├── frontmatter/ (get, set, transition)
    ├── chain/       (resolve)
    ├── index/       (build, query, files, status)
    ├── git/         (context, changed-files, sensitive-check)
    ├── spec/        (coverage)
    └── template/    (scaffold)
```

The MCP server and the Cobra CLI are parallel frontends to the same internal packages.

## Tool Mapping

| MCP Tool | CLI Equivalent | Input Fields |
|---|---|---|
| `rpi_scan` | `rpi scan` | `type?, status?, proposal?, references?, archivable?` |
| `rpi_scaffold` | `rpi scaffold <type>` | `type, topic, ticket?, research?, proposal?, spec?, tags?, write?, force?` |
| `rpi_frontmatter_get` | `rpi frontmatter get` | `file, field?` |
| `rpi_frontmatter_set` | `rpi frontmatter set` | `file, field, value` |
| `rpi_frontmatter_transition` | `rpi frontmatter transition` | `file, status` |
| `rpi_chain` | `rpi chain` | `path, sections?` |
| `rpi_extract` | `rpi extract --section` | `path, section` |
| `rpi_extract_list_sections` | `rpi extract --list-sections` | `path` |
| `rpi_verify_completeness` | `rpi verify completeness` | `plan_path` |
| `rpi_verify_markers` | `rpi verify markers` | `file_path?` |
| `rpi_index_build` | `rpi index build` | `lang?, path?, force?` |
| `rpi_index_query` | `rpi index query` | `pattern, kind?, exported?` |
| `rpi_index_files` | `rpi index files` | `lang?` |
| `rpi_index_status` | `rpi index status` | (none) |
| `rpi_git_context` | `rpi git-context` | (none) |
| `rpi_git_changed_files` | `rpi git-context changed-files` | (none) |
| `rpi_git_sensitive_check` | `rpi git-context sensitive-check` | (none) |
| `rpi_archive_scan` | `rpi archive scan` | (none) |
| `rpi_archive_check_refs` | `rpi archive check-refs` | `path` |
| `rpi_archive_move` | `rpi archive move` | `path, force?` |
| `rpi_spec_coverage` | `rpi spec coverage` | `spec_file` |

## File Structure

```
cmd/rpi/
├── serve.go          # NEW: rpi serve subcommand + all 20 tool registrations
├── serve_test.go     # NEW: tests for tool handlers
├── main.go           # MODIFIED: register serveCmd (one line in init)
└── (all other files unchanged)
```

New dependency in `go.mod`:
```
github.com/modelcontextprotocol/go-sdk
```

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| SDK breaking changes (recent library) | Medium | Medium | Pin version in go.mod; SDK is officially maintained |
| 20 tools clutter MCP client UI | Low | Low | Clear `rpi_` prefix + descriptions; can group later |
| Dual CLI/MCP maintenance burden | Low | Low | Both call same internal packages; logic changes in one place |

## What This Proposal Does NOT Cover

- Full portability of skill/command files across AI clients (separate effort)
- SSE/HTTP transport (stdio is sufficient for local tool servers)
- Authentication or multi-user scenarios
- Deprecation of the CLI interface

## Open Questions

None.

## References

- [Official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Specification](https://modelcontextprotocol.io)
- Current CLI reference: `.rpi/cli-reference.md`
