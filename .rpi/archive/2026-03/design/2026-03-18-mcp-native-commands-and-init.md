---
date: "2026-03-18"
status: complete
tags:
    - mcp
    - init
    - commands
    - skills
topic: mcp-native-commands-and-init
---

# Proposal: MCP-Native Commands and Init

## Summary

Update `rpi init` to auto-configure the MCP server, and rewrite all commands and skills to reference MCP tool names directly instead of shelling out to the `rpi` CLI. One `rpi init` gives the AI tool everything it needs: slash commands that orchestrate workflows using typed MCP tools.

## Investigation Findings

Three layers exist but don't connect:

| Layer | Location | Count | Role |
|-------|----------|-------|------|
| Commands | `internal/workflow/assets/commands/` | 7 | High-level workflow orchestration |
| Skills | `internal/workflow/assets/skills/` | 4 | Discovery & analysis helpers |
| MCP Server | `cmd/rpi/serve.go` | 21 tools | Typed low-level operations |

Commands tell Claude to shell out: "Use `rpi` to scaffold..." causing Claude to construct shell commands. The MCP server is fully implemented (21 tools with typed schemas in `serve.go:69-178`) but never configured during init, so it sits unused.

`rpi init` (`init_cmd.go:120-240`) creates dirs, installs workflow files, builds index, generates CLI reference — but writes no MCP configuration.

## Constraints & Requirements

### Must
- `rpi init` configures MCP server for Claude Code (`.claude/settings.local.json`)
- All 7 commands reference MCP tool names (e.g., `rpi_scaffold`) instead of CLI
- All 4 skills reference MCP tool names where they currently use CLI
- `rpi init` checks if `rpi` is in PATH before writing MCP config
- Existing `settings.local.json` is merged, not overwritten
- User is warned if `settings.local.json` already has an `rpi` MCP server entry
- `--no-mcp` flag skips MCP configuration

### Must Not
- Break existing `rpi init` behavior (dirs, templates, index, CLI reference)
- Auto-configure MCP for OpenCode (deferred)
- Modify the MCP server implementation itself

### Out of Scope
- OpenCode MCP configuration
- New MCP tools beyond the existing 21
- `.thoughts/` to `.rpi/` migration in locally installed files

## Design Decisions

### DD-1: MCP configuration in init

**Chosen**: Auto-configure by writing to `.claude/settings.local.json` with merge semantics.

**Alternatives considered**: Print instructions only (non-invasive but defeats "easy init" goal).

**Rationale**: If `rpi init` creates slash commands but doesn't configure MCP, the commands can't reach their tools. Using `settings.local.json` (not `settings.json`) keeps the config local — different machines may have `rpi` at different paths.

**How it works**:
1. Check `rpi` is in PATH via `exec.LookPath("rpi")` — if not, warn and skip MCP config
2. Read existing `.claude/settings.local.json` if present
3. If `mcpServers.rpi` already exists, warn ("MCP server 'rpi' already configured") but don't overwrite
4. Otherwise, merge `{"mcpServers": {"rpi": {"command": "rpi", "args": ["serve"]}}}` into existing JSON
5. Write back with proper formatting

### DD-2: Commands become MCP-native

**Chosen**: Rewrite all 7 command files to reference MCP tool names directly.

**Alternatives considered**: Keep tool-agnostic "Use `rpi` to..." phrasing (Claude often shells out when MCP is available); add MCP hints in parentheses (verbose, cluttered).

**Rationale**: If `rpi init` configures MCP, commands and MCP tools always coexist. Using explicit tool names is unambiguous — Claude knows exactly which tool to call.

**How it works**: Mapping of operations to MCP tool names:

| Operation | MCP Tool |
|-----------|----------|
| Scaffold artifact | `rpi_scaffold` |
| Read frontmatter | `rpi_frontmatter_get` |
| Set frontmatter field | `rpi_frontmatter_set` |
| Transition status | `rpi_frontmatter_transition` |
| Resolve chain | `rpi_chain` |
| Scan artifacts | `rpi_scan` |
| Extract section | `rpi_extract` |
| List sections | `rpi_extract_list_sections` |
| Check completeness | `rpi_verify_completeness` |
| Scan markers | `rpi_verify_markers` |
| Query index | `rpi_index_query` |
| Build index | `rpi_index_build` |
| Index status | `rpi_index_status` |
| List indexed files | `rpi_index_files` |
| Git context | `rpi_git_context` |
| Changed files | `rpi_git_changed_files` |
| Sensitive check | `rpi_git_sensitive_check` |
| Check refs | `rpi_archive_check_refs` |
| Archive move | `rpi_archive_move` |
| Archive scan | `rpi_archive_scan` |
| Spec coverage | `rpi_spec_coverage` |

### DD-3: Skills updated for consistency

**Chosen**: Update all 4 skills to reference MCP tool names.

**Skills affected**:
- `locate-codebase` — `rpi index status/query/files` → `rpi_index_status/query/files`
- `locate-thoughts` — `rpi scan` → `rpi_scan`
- `analyze-thoughts` — `rpi extract/frontmatter` → `rpi_extract/frontmatter_get`
- `find-patterns` — `rpi index query` → `rpi_index_query`

### DD-4: Init flow ordering

**Chosen**: MCP config step added after workflow file installation, before index building.

**Updated flow**:
```
rpi init
  1. Create tool dirs (.claude/agents, commands, skills, hooks)
  2. Create .rpi/ artifact dirs
  3. Generate rules file (CLAUDE.md)
  4. Generate PIPELINE.md
  5. Update .gitignore
  6. Install embedded workflow files
  7. [NEW] Configure MCP server in .claude/settings.local.json
  8. Build codebase index
  9. Generate CLI reference
```

## File Structure

Modified files:
- `cmd/rpi/init_cmd.go` — add `configureMCP()` function and `--no-mcp` flag
- `cmd/rpi/init_cmd_test.go` — test MCP configuration
- `internal/workflow/assets/commands/rpi-research.md` — MCP tool refs
- `internal/workflow/assets/commands/rpi-propose.md` — MCP tool refs
- `internal/workflow/assets/commands/rpi-plan.md` — MCP tool refs
- `internal/workflow/assets/commands/rpi-implement.md` — MCP tool refs
- `internal/workflow/assets/commands/rpi-verify.md` — MCP tool refs
- `internal/workflow/assets/commands/rpi-archive.md` — MCP tool refs
- `internal/workflow/assets/commands/rpi-commit.md` — MCP tool refs
- `internal/workflow/assets/skills/locate-codebase/SKILL.md` — MCP tool refs
- `internal/workflow/assets/skills/locate-thoughts/SKILL.md` — MCP tool refs
- `internal/workflow/assets/skills/analyze-thoughts/SKILL.md` — MCP tool refs
- `internal/workflow/assets/skills/find-patterns/SKILL.md` — MCP tool refs
- `internal/workflow/assets/templates/PIPELINE.md.template` — MCP tool refs

## Risks & Mitigations

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Existing `settings.local.json` corruption | Low | Parse-merge-write with JSON validation; warn on existing `rpi` entry |
| `rpi` not in PATH at runtime | Medium | Init checks PATH and warns; MCP server fails gracefully |
| MCP tool names change | Very Low | Names are stable; changing them breaks all MCP clients |

## What This Proposal Does NOT Cover

- OpenCode MCP configuration (separate proposal)
- New MCP tools or server changes
- Changes to CLAUDE.md template content beyond tool references
- `rpi init --update` MCP re-configuration

## Open Questions

None — all resolved during discussion.

## References

- `cmd/rpi/serve.go` — MCP server with 21 tools
- `cmd/rpi/init_cmd.go` — current init implementation
- `internal/workflow/workflow.go` — embedded asset installation
- `internal/workflow/assets/` — source of truth for workflow files
