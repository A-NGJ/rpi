---
date: 2026-04-07T16:00:00+02:00
related_research: .rpi/research/2026-04-07-rpi-improvements-via-claude-code-internals.md
spec: .rpi/specs/agent-skills.md
status: complete
tags:
    - design
    - skills
    - metadata
    - safety
topic: skill metadata for read-only safety
---

# Design: Skill Metadata for Read-Only Safety

## Summary

Enhance the Claude Code integration by adding `allowed-tools` and `context` metadata to skill frontmatter during `rpi init`/`rpi update`. Read-only skills (research, verify, explain) get restricted to non-writing tools. The research skill gets `context: fork` for isolated exploration that doesn't pollute the main conversation.

## Context

The research (.rpi/research/2026-04-07-rpi-improvements-via-claude-code-internals.md) identified skill metadata as underutilized. The agent-skills spec explicitly deferred `allowed-tools` as out of scope. This design picks it up.

Currently, all 10 RPI skills can use any Claude Code tool — a research skill could accidentally write files, and a verify skill could edit code. The `skillOverrides` mechanism in `workflow.go:32` already injects tool-specific frontmatter (model, disable-model-invocation) for 2 skills. Extending it to inject `allowed-tools` and `context` for 3 more skills is the same pattern.

## Constraints

- Metadata injected at install time via `skillOverrides` — canonical SKILL.md files stay clean
- Only applies to Claude and OpenCode targets — `agents-only` gets no metadata (existing behavior)
- Must not break existing skills or require changes to SKILL.md content
- The `allowed-tools` list must include MCP tools the skills use (prefixed `mcp__rpi__`)

## Design

### Skills Receiving Metadata

| Skill | `allowed-tools` | `context` | Reasoning |
|---|---|---|---|
| `rpi-research` | Read, Glob, Grep, Bash, Agent, WebSearch, WebFetch, + all MCP tools | `fork` | Investigation only — never writes files. Fork prevents context pollution from extensive file reading. |
| `rpi-verify` | Read, Glob, Grep, Bash, Agent, + all MCP tools | (none) | Verification only — never writes files. Stays in main context so it can see recent implementation work. |
| `rpi-explain` | Read, Glob, Grep, Bash, Agent, + all MCP tools | (none) | Explanation only — never writes files. Stays in main context for diff-scoped walkthrough. |

Skills NOT receiving metadata (they need Write/Edit): `rpi-implement`, `rpi-plan`, `rpi-propose`, `rpi-diagnose`, `rpi-spec-sync`, `rpi-commit`, `rpi-archive` (already has model override).

### MCP Tool Enumeration

The `allowed-tools` list must explicitly name each MCP tool. Current MCP tools (20):

```
mcp__rpi__rpi_git_context, mcp__rpi__rpi_git_changed_files,
mcp__rpi__rpi_git_sensitive_check, mcp__rpi__rpi_archive_scan,
mcp__rpi__rpi_scan, mcp__rpi__rpi_scaffold, mcp__rpi__rpi_frontmatter_get,
mcp__rpi__rpi_frontmatter_set, mcp__rpi__rpi_frontmatter_transition,
mcp__rpi__rpi_chain, mcp__rpi__rpi_extract,
mcp__rpi__rpi_extract_list_sections, mcp__rpi__rpi_verify_completeness,
mcp__rpi__rpi_verify_markers, mcp__rpi__rpi_verify_spec,
mcp__rpi__rpi_context_essentials, mcp__rpi__rpi_session_resume,
mcp__rpi__rpi_suggest_next, mcp__rpi__rpi_archive_check_refs,
mcp__rpi__rpi_archive_move
```

The research/verify/explain skills won't call all of these, but including all prevents breakage if skill prompts evolve. Write-capable MCP tools (scaffold, frontmatter_set, frontmatter_transition, archive_move) are included because they operate on `.rpi/` artifacts, not source code — the safety boundary is about source file writes via Write/Edit, not artifact metadata.

### Implementation

The `skillOverrides` map in `internal/workflow/workflow.go:32` gains 3 new entries. The `allowed-tools` value is a comma-separated string that `injectFrontmatter` writes as a single YAML field. Claude Code parses skill frontmatter and interprets `allowed-tools` as a tool restriction list.

**Updated `skillOverrides`**:

```go
var skillOverrides = map[string]map[string]string{
    "rpi-archive":  {"model": "haiku", "disable-model-invocation": "true"},
    "rpi-commit":   {"model": "haiku"},
    "rpi-research": {"allowed-tools": readOnlyTools + ",WebSearch,WebFetch", "context": "fork"},
    "rpi-verify":   {"allowed-tools": readOnlyTools},
    "rpi-explain":  {"allowed-tools": readOnlyTools},
}
```

Where `readOnlyTools` is a const string containing the base read-only tool set plus all MCP tools.

### `injectFrontmatter` Changes

The existing function writes fields as `key: value` lines. For `allowed-tools`, the value is a comma-separated string. In YAML this renders as:

```yaml
allowed-tools: Read,Glob,Grep,Bash,Agent,mcp__rpi__rpi_scan,...
```

No changes needed to `injectFrontmatter` — it already handles arbitrary string values.

### Effect on `rpi update`

Users with existing installs get the new metadata after running `rpi update --force`. Without `--force`, existing SKILL.md files are not overwritten (existing behavior).

## Alternatives Considered

- **Deny-list instead of allow-list**: Block only Write/Edit instead of listing all allowed tools. Claude Code's `allowed-tools` is an allow-list, not a deny-list. Not an option.
- **`context: fork` for verify and explain**: Rejected — these skills benefit from seeing recent implementation context. Only research is clearly self-contained.
- **No MCP tools in allow-list**: Would break all MCP tool calls from read-only skills. Must include them.

## Risks

- **Claude Code `allowed-tools` behavior**: If Claude Code doesn't support `allowed-tools` in user-defined skill frontmatter (only in bundled skills), this has no effect. Mitigation: the fields are inert if unrecognized — no breakage, just no enforcement.
- **MCP tool list gets stale**: When new MCP tools are added, the `readOnlyTools` const must be updated. Mitigation: define the list in one place; the serve.go registration is the source of truth.
- **`context: fork` interaction with research arguments**: If the user discusses a topic and then invokes `/rpi-research`, the forked context won't see the discussion. Mitigation: the skill takes explicit arguments; the user passes the topic.

## Out of Scope

- Per-skill hooks (e.g., PreToolUse for plan boundary enforcement)
- `agent` field (designating a custom agent for a skill)
- `files` field (bundling reference files with a skill)
- New init targets (cursor, generic)
- Plugin packaging

## References

- Research: .rpi/research/2026-04-07-rpi-improvements-via-claude-code-internals.md
- Existing spec: .rpi/specs/agent-skills.md (removing `allowed-tools` from Out of Scope)
