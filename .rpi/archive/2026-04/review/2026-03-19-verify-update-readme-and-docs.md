---
archived_date: "2026-04-04"
date: 2026-03-19T13:08:56+01:00
plan: .rpi/plans/2026-03-19-update-readme-and-docs.md
status: archived
tags:
    - verify
topic: update readme and docs
---

# Verification Report: Update README and Docs

## Summary

**Status: Pass with warnings**

- Blockers: 0
- Warnings: 2
- Notes: 1

Two-phase documentation update: (1) renamed all `proposals/` references to `designs/` across 4 doc files, and (2) added MCP server documentation to README, rpi-init, and architecture docs. Both phases committed.

## Completeness

**Plan checkboxes**: 4/12 checked. The 8 unchecked items are manual verification steps and commit steps that were completed but not ticked in the plan file. All actual work was done.

**File coverage**: All 6 planned files were modified:
- `README.md` -- proposals→designs rename + MCP section + init bullet
- `docs/thoughts-directory.md` -- proposals→designs rename
- `docs/stages.md` -- proposals→designs rename
- `docs/workflow-guide.md` -- proposals→designs rename (5 locations)
- `docs/rpi-init.md` -- `--no-mcp` flag + MCP config section + proposals→designs fix
- `docs/architecture.md` -- MCP server bullet + project structure update

**Grep verification**: `grep -r "proposals/" README.md docs/` returns zero matches. All stale references eliminated.

**Markers**: No new TODO/FIXME/HACK markers introduced. The 9 markers found are all pre-existing (template placeholders, serve.go string literals, command descriptions).

## Correctness

### Phase 1: proposals→designs rename

All replacements are correct and read naturally in context:
- `docs/thoughts-directory.md:11` -- directory tree updated
- `docs/thoughts-directory.md:59` -- "Design documents preserve decision rationale" reads well
- `docs/stages.md:27` -- "documenting the design in `.rpi/designs/`" correct
- `docs/stages.md:92` -- `archive/designs/` correct
- `docs/workflow-guide.md:49,53,88,92,125` -- all path references and prose updated
- `README.md:8,12-13,73` -- diagram and table updated

### Phase 2: MCP documentation

- **README.md:59** -- MCP init bullet accurately describes auto-registration behavior
- **README.md:100-104** -- MCP Server section is concise (4 lines), mentions `rpi serve`, tool names, auto-registration, `--no-mcp`, and links to architecture.md
- **docs/rpi-init.md:21** -- `--no-mcp` flag listed in options block
- **docs/rpi-init.md:35** -- MCP registration listed under Claude Code target
- **docs/rpi-init.md:52-60** -- MCP Server Configuration subsection covers all behaviors from the spec (MC-1 through MC-6, MC-12): PATH requirements, skip conditions, warn-don't-fail, registration command, `--update` exclusion
- **docs/architecture.md:16** -- MCP server bullet matches serve.go implementation
- **docs/architecture.md:24** -- Project structure comment updated

**Spec cross-check** (`.rpi/specs/mcp-init-and-commands.md`):
- MC-1 (init writes MCP config): documented in rpi-init.md
- MC-2 (warn if rpi not in PATH): documented ("Warns and continues")
- MC-5 (`--no-mcp` skips): documented in options and MCP section
- MC-6 (config format): documented (`claude mcp add rpi -- rpi serve`)

**Spec cross-check** (`.rpi/specs/mcp-server.md`):
- MS-1 (`rpi serve` on stdio): documented in README and architecture
- MS-3/MS-4 (tool registration/naming): documented with example tool names

## Coherence

- Writing style matches existing docs (double-dash em dashes, backtick formatting for commands and paths)
- MCP section placement in README (after Documentation, before How It Compares) fits the narrative flow
- rpi-init.md MCP subsection follows the same pattern as existing subsections (Claude Code target, OpenCode target, Shared)
- architecture.md bullet list order is logical (MCP server last, as it wraps all other capabilities)

## Issues

### Blockers

None.

### Warnings

1. **README.md:77** -- `/rpi-verify` description still says "matches proposal artifacts" rather than "matches design artifacts". Not introduced by this change (pre-existing), but now inconsistent with the rename.

2. **docs/rpi-init.md:35** -- Documents `claude mcp add rpi -- rpi serve` but the actual implementation in `init_cmd.go:336` uses this exact command. However, the spec (MC-6) originally described a `settings.local.json` approach which was later changed to `claude mcp add`. The docs match the implementation, but if the spec is considered authoritative, there's a discrepancy. The spec should be updated.

### Notes

1. **Plan checkboxes not fully ticked** -- 8 unchecked items remain in `.rpi/plans/2026-03-19-update-readme-and-docs.md` (manual verification and commit steps). Cosmetic only; all work was completed and committed.

## References

- Plan: `.rpi/plans/2026-03-19-update-readme-and-docs.md`
- MCP server spec: `.rpi/specs/mcp-server.md`
- MCP init spec: `.rpi/specs/mcp-init-and-commands.md`
- Commits: `993a599`, `3c3269b`
