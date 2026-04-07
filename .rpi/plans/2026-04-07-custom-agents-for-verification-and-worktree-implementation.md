---
date: 2026-04-07T16:32:55+02:00
design: .rpi/designs/2026-04-07-custom-agents-for-verification-and-worktree-implementation.md
spec: .rpi/specs/custom-agents.md
status: complete
tags:
    - plan
topic: custom agents for verification and worktree implementation
---

# Custom Agents for Verification and Worktree Implementation — Implementation Plan

## Overview

Add two Claude Code custom agent definitions and wire them into the `rpi init`/`rpi update` installation flow. Update the `/rpi-implement` skill prompt to support worktree mode.

**Scope**: 4 files modified, 2 new files

## Source Documents
- **Design**: .rpi/designs/2026-04-07-custom-agents-for-verification-and-worktree-implementation.md
- **Spec**: .rpi/specs/custom-agents.md

## Phase 1: Agent Definition Files

### Overview
Create the two agent markdown files as embedded assets in the Go binary.

### Tasks:

#### 1. Verification Agent Definition
**File**: `internal/workflow/assets/agents/rpi-verify.md`
**Changes**: New file. Markdown with YAML frontmatter (`name`, `description`, `allowed-tools`). Prompt instructs the agent to: resolve artifact chain from a given plan/spec path, extract scenarios via the verify spec tool, check each scenario against actual code with file:line references, run completeness and marker checks, return structured pass/fail summary with overall verdict.

#### 2. Worktree Implementation Agent Definition
**File**: `internal/workflow/assets/agents/rpi-implement-worktree.md`
**Changes**: New file. Markdown with YAML frontmatter (`name`, `description`). No tool restrictions (needs Write/Edit/Bash). Prompt instructs the agent to: implement all remaining plan phases from a provided context bundle, run tests/linters after each phase, commit after each passing phase with descriptive messages, update plan checkboxes, return structured summary of phases completed/commits made/issues encountered.

### Success Criteria:

#### Automated Verification:
- [x] `go build ./...` passes (embedded assets picked up by `//go:embed all:assets`)
- [x] Both files parseable as markdown with valid YAML frontmatter

### Commit:
- [x] Stage: `internal/workflow/assets/agents/`
- [x] Message: `feat(agents): add verification and worktree implementation agent definitions`

---

## Phase 2: InstallAgents Function

### Overview
Add `InstallAgents` to `workflow.go` mirroring `InstallSkills` but simpler (no frontmatter injection). Add tests.

### Tasks:

#### 1. InstallAgents Function
**File**: `internal/workflow/workflow.go`
**Changes**: Add `InstallAgents(agentsDir string, force bool) (int, error)`. Reads embedded files from `assets/agents/`, copies each `.md` file to `<agentsDir>/<name>.md`. Only overwrites existing files when `force=true`. Returns count of files written.

#### 2. Tests
**File**: `internal/workflow/workflow_test.go`
**Changes**: Add tests:
- `TestInstallAgents_InstallsAllAgents`: verifies both agent files are installed to the target directory
- `TestInstallAgents_NoOverwriteWithoutForce`: verifies existing files are preserved without `--force`
- `TestInstallAgents_ForceOverwrites`: verifies `--force` replaces existing files
- `TestAgentDefinitions_ValidFrontmatter`: verifies each agent `.md` has `name` and `description` frontmatter fields

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/workflow/...` passes with new tests

### Commit:
- [x] Stage: `internal/workflow/workflow.go`, `internal/workflow/workflow_test.go`
- [x] Message: `feat(agents): add InstallAgents function for agent definition installation`

---

## Phase 3: Wire into Init/Update Flow

### Overview
Add `agents` to Claude target subdirs, call `InstallAgents` from `syncProject`, update existing tests.

### Tasks:

#### 1. Add Agents Subdir to Claude Target
**File**: `cmd/rpi/init_cmd.go`
**Changes**: In `resolveTargetConfig("claude")`, add `"agents"` to the `subdirs` slice: `[]string{"skills", "hooks", "agents"}`. OpenCode and agents-only remain unchanged.

#### 2. Call InstallAgents from syncProject
**File**: `cmd/rpi/sync.go`
**Changes**: After the `InstallSkills` call, add a conditional block: if `target == TargetClaude`, compute `agentsDir = filepath.Join(targetDir, cfg.toolDir, "agents")` and call `workflow.InstallAgents(agentsDir, opts.force)`. Log the count.

#### 3. Update Init Tests
**File**: `cmd/rpi/init_cmd_test.go`
**Changes**:
- `TestInitCreatesAllDirs`: Remove AS-14 assertion that `.claude/agents/` does NOT exist. Add assertion that `.claude/agents/` IS created and contains `rpi-verify.md` and `rpi-implement-worktree.md`.
- `TestInitOpenCode`: Keep assertion that `.opencode/agents/` does NOT exist (agents are Claude-only).
- `TestInitAgentsOnly`: Keep assertion that no agent definitions are installed.
- Add `TestInitInstallsAgents`: Verify both agent files exist after `rpi init` with Claude target.

#### 4. Update Update Tests
**File**: `cmd/rpi/update_cmd_test.go`
**Changes**: Add `TestUpdateSyncsAgents`: Init, delete an agent file, run update, verify it's restored. Add `TestUpdateAgentsOnlyNoAgents`: Verify agents-only update does not create agent files.

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/...` passes (all existing + new tests)
- [x] `go test ./...` passes (full suite)

### Commit:
- [x] Stage: `cmd/rpi/init_cmd.go`, `cmd/rpi/sync.go`, `cmd/rpi/init_cmd_test.go`, `cmd/rpi/update_cmd_test.go`
- [x] Message: `feat(agents): wire agent installation into init and update commands`

---

## Phase 4: Update Implement Skill Prompt

### Overview
Add worktree mode instructions to the `/rpi-implement` skill.

### Tasks:

#### 1. Add Worktree Mode Section
**File**: `internal/workflow/assets/skills/rpi-implement/SKILL.md`
**Changes**: Add a new section after Principles:

```
## Worktree Mode

If the worktree implementation agent is available:
- After pre-review approval, spawn the agent in a worktree with the full context bundle (plan content, spec scenarios, design constraints, file paths to read)
- On agent completion, spawn the verification agent to check the worktree branch
- If verification passes and no manual verification items exist in the plan, merge the worktree branch to the current branch automatically
- If manual verification items exist, present the diff and wait for user approval before merging
- After merge, update plan status to complete

If agents are not available, implement in-place on the current branch (default behavior above).
```

Note: the prompt must not reference MCP tool names (enforced by `TestPromptStructure_NoToolReferences`) or internal file paths. Keep it behavioral.

#### 2. Verify Prompt Constraints
**Changes**: Ensure the updated SKILL.md stays under 50 lines (body excluding frontmatter, enforced by `TestPromptStructure_LineCount`) and contains no `rpi_` or backtick-quoted `rpi ` references.

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/workflow/...` passes (prompt structure tests: line count, no tool refs, required sections)

### Commit:
- [x] Stage: `internal/workflow/assets/skills/rpi-implement/SKILL.md`
- [x] Message: `feat(agents): add worktree mode instructions to implement skill`

---

## References
- Design: .rpi/designs/2026-04-07-custom-agents-for-verification-and-worktree-implementation.md
- Spec: .rpi/specs/custom-agents.md
