---
archived_date: "2026-04-04"
date: 2026-03-19T13:02:45+01:00
status: archived
tags:
    - plan
topic: update readme and docs
---

# Update README and Docs — Implementation Plan

## Overview

Two concerns: (1) rename stale `proposals/` references to `designs/` across all docs, and (2) document the MCP server and MCP auto-configuration that were added recently but never reflected in documentation.

**Scope**: 5 files modified, 0 new files

## Phase 1: Rename `proposals/` → `designs/` across docs

### Overview

All Go code and commands already use `designs/`. The docs still reference the old `proposals/` directory name.

### Tasks:

#### 1. README.md
**File**: `README.md`
**Changes**:
- Line 8: update ASCII diagram `Propose` → `Design` in header
- Line 12: `proposals/` → `designs/`
- Line 13: add `specs/` to diagram (currently missing)
- Line 72: output path `.rpi/proposals/YYYY-MM-DD-topic.md` → `.rpi/designs/YYYY-MM-DD-topic.md`
- Line 72: description update "proposes solutions with trade-offs" → "Investigates, analyzes, and designs solutions with trade-offs"

#### 2. docs/thoughts-directory.md
**File**: `docs/thoughts-directory.md`
**Changes**:
- Line 11: `proposals/` → `designs/` in directory tree + comment
- Line 59: `.rpi/proposals/` → `.rpi/designs/` in explanation text; "Proposal documents" → "Design documents"

#### 3. docs/stages.md
**File**: `docs/stages.md`
**Changes**:
- Line 27: `documenting the proposal in .rpi/proposals/` → `documenting the design in .rpi/designs/`
- Line 92: `archive/proposals/` → `archive/designs/`

#### 4. docs/workflow-guide.md
**File**: `docs/workflow-guide.md`
**Changes**:
- Line 49: `writes the proposal...to .rpi/proposals/` → `writes the design...to .rpi/designs/`
- Line 53: `.rpi/proposals/2026-03-04-api-rate-limiting.md` → `.rpi/designs/...`
- Line 88: `writes the full proposal to .rpi/proposals/` → `writes the full design to .rpi/designs/`
- Line 92: `.rpi/proposals/2026-03-04-notification-system.md` → `.rpi/designs/...`
- Line 125: `research/proposals/plan` → `research/designs/plan`

### Success Criteria:

#### Automated Verification:
- [x] `grep -r "proposals/" README.md docs/` returns zero matches
- [x] `go test ./...` passes (no Go code changed, but sanity check)

#### Manual Verification:
- [ ] Skim each file to confirm replacements read naturally in context

### Commit:
- [ ] Stage: `README.md`, `docs/thoughts-directory.md`, `docs/stages.md`, `docs/workflow-guide.md`
- [ ] Message: `docs: rename proposals to designs across all documentation`

**Note**: Pause for manual confirmation before proceeding to next phase.

---

## Phase 2: Document MCP server and init auto-configuration

### Overview

The project now has an MCP server (`rpi serve`) exposing 21 tools and `rpi init` auto-configures it. None of this is reflected in README, rpi-init docs, or architecture docs.

### Tasks:

#### 1. README.md — Add MCP section and update init output
**File**: `README.md`
**Changes**:
- In "What `rpi init` creates" list (after line 58): add bullet for MCP auto-configuration via `claude mcp add`
- Add `--no-mcp` to the init command example or mention it
- After the "How It Compares" section (or before it), add a short "MCP Server" section explaining:
  - `rpi serve` starts an MCP server over stdio
  - `rpi init` auto-registers it with Claude Code
  - AI tools call typed MCP tools (`rpi_scaffold`, `rpi_scan`, etc.) instead of shelling out
  - Link to architecture.md for details

#### 2. docs/rpi-init.md — Document MCP auto-configuration
**File**: `docs/rpi-init.md`
**Changes**:
- Add `--no-mcp` to the Options list (line 21 area)
- Under "Claude Code target" section: add bullet for MCP server registration (`claude mcp add rpi -- rpi serve`)
- Add a subsection "MCP Server Configuration" explaining:
  - Auto-registered when `rpi` and `claude` are both in PATH
  - Skipped with `--no-mcp` or when target is opencode
  - Warns (doesn't fail) if `rpi`/`claude` not in PATH or already configured
- Fix `proposals` reference in line 31 ("proposals" in templates description) → `designs`

#### 3. docs/architecture.md — Add MCP server to architecture
**File**: `docs/architecture.md`
**Changes**:
- Add "MCP Server" bullet to the "The binary handles" list, explaining `rpi serve` exposes all CLI operations as typed MCP tools over stdio so the LLM calls validated schemas instead of constructing shell commands
- Update project structure tree: add `cmd/rpi/serve.go` and `cmd/rpi/serve_test.go`

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` passes
- [x] `grep -r "proposals/" docs/rpi-init.md` returns zero matches

#### Manual Verification:
- [ ] README MCP section is concise (no longer than the existing "How It Compares" section)
- [ ] rpi-init.md accurately reflects the `--no-mcp` flag behavior
- [ ] architecture.md MCP description matches the actual serve.go implementation

### Commit:
- [ ] Stage: `README.md`, `docs/rpi-init.md`, `docs/architecture.md`
- [ ] Message: `docs: add MCP server and init auto-configuration to documentation`

---

## References

- `.rpi/specs/mcp-server.md` — MCP server behavioral spec
- `.rpi/specs/mcp-init.md` — MCP init and commands spec
- `cmd/rpi/serve.go` — MCP server implementation
- `cmd/rpi/init_cmd.go` — Init with MCP auto-configuration
