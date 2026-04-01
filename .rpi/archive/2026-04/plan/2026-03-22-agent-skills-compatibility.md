---
archived_date: "2026-04-02"
date: 2026-03-22T23:32:57+01:00
design: .rpi/designs/2026-03-22-agent-skills-compatibility.md
spec: .rpi/specs/agent-skills-compatibility.md
status: archived
tags:
    - plan
topic: agent-skills-compatibility
---

# Agent Skills Compatibility — Implementation Plan

## Overview

Restructure all 14 RPI workflow files into Agent Skills-compliant `skills/*/SKILL.md` format. Canonical skills have only `name` + `description` in frontmatter. Tool-specific fields (`model`, `tools`, `disable-model-invocation`) are injected programmatically at install time. Skills install to `.agents/skills/` (cross-tool) plus the active tool directory.

**Scope**: ~6 files modified, 14 asset files restructured, ~3 test files updated

## Source Documents
- **Design**: .rpi/designs/2026-03-22-agent-skills-compatibility.md
- **Spec**: .rpi/specs/agent-skills-compatibility.md

---

## Phase 1: Restructure embedded assets

### Overview
Convert all commands and the agent into `skills/*/SKILL.md` directories. Strip tool-specific frontmatter fields from canonical copies. Remove `assets/commands/` and `assets/agents/`.

### Tasks:

#### 1. Convert 9 commands to skill directories
**Files**: `internal/workflow/assets/skills/{rpi-research,rpi-propose,rpi-plan,rpi-implement,rpi-verify,rpi-diagnose,rpi-explain,rpi-commit,rpi-archive}/SKILL.md`
**Changes**:
- Create a `<name>/SKILL.md` for each command
- Frontmatter: `name: <directory-name>` + `description: <existing description>` (no `model`, no `disable-model-invocation`)
- Body: identical to current command markdown body

#### 2. Convert agent to skill directory
**File**: `internal/workflow/assets/skills/codebase-analyzer/SKILL.md`
**Changes**:
- Move content from `assets/agents/codebase-analyzer.md`
- Frontmatter: `name: codebase-analyzer` + `description: <existing>` (no `model`, no `tools`)
- Body: identical

#### 3. Strip tool-specific fields from existing 4 skills
**Files**: `internal/workflow/assets/skills/{find-patterns,analyze-thoughts,locate-codebase,locate-thoughts}/SKILL.md`
**Changes**: Remove `model:` from frontmatter. Keep `name` and `description` only.

#### 4. Remove old directories
**Delete**: `internal/workflow/assets/commands/` (9 files), `internal/workflow/assets/agents/` (1 file)

### Success Criteria:

#### Automated Verification:
- [x] `ls internal/workflow/assets/skills/` shows exactly 13 directories (codebase-analyzer dropped — agent-specific, not portable)
- [x] Every `SKILL.md` has only `name` and `description` in frontmatter (no `model`, `tools`, `disable-model-invocation`)
- [x] Every `name` field matches its parent directory name
- [x] `internal/workflow/assets/commands/` and `internal/workflow/assets/agents/` do not exist
- [x] `go build ./...` succeeds

### Commit:
- [ ] Stage: `internal/workflow/assets/`
- [ ] Message: `refactor: convert all workflow files to Agent Skills format`

---

## Phase 2: Rewrite `workflow.go` install logic

### Overview
Replace the walk-and-transform approach with dual-directory install. Canonical skills go to `.agents/skills/`, tool-enriched copies go to `<toolDir>/skills/`. Tool-specific frontmatter is injected from a static map.

### Tasks:

#### 1. Add `TargetAgentsOnly` and skill model map
**File**: `internal/workflow/workflow.go`
**Changes**:
- Add `TargetAgentsOnly Target = "agents-only"`
- Add a `skillOverrides` map defining per-skill tool-specific fields:
  ```go
  var skillOverrides = map[string]map[string]string{
      "rpi-archive":       {"model": "haiku", "disable-model-invocation": "true"},
      "rpi-commit":        {"model": "haiku"},
      "find-patterns":     {"model": "sonnet"},
      "analyze-thoughts":  {"model": "sonnet"},
      "locate-codebase":   {"model": "haiku"},
      "locate-thoughts":   {"model": "haiku"},
      "codebase-analyzer": {"model": "inherit"},
      // remaining 7 pipeline skills: model=inherit (Claude default, no injection needed)
  }
  ```

#### 2. Rewrite `InstallTo` → `InstallSkills`
**File**: `internal/workflow/workflow.go`
**Changes**:
- New signature: `InstallSkills(agentsDir, toolDir string, target Target, force bool) (int, error)`
- Walk `assets/skills/` subdirectories
- For every skill:
  - Install canonical SKILL.md → `agentsDir/<skill>/SKILL.md` (always)
  - If target != `TargetAgentsOnly`: install enriched copy → `toolDir/skills/<skill>/SKILL.md` with injected frontmatter fields from `skillOverrides` + OpenCode model transforms
- Skip `assets/templates/` (not skills)
- Return total files installed

#### 3. Add `injectFrontmatter` helper
**File**: `internal/workflow/workflow.go`
**Changes**: Function that takes SKILL.md content + field map, inserts fields into existing YAML frontmatter block. For OpenCode target, also translate model aliases via existing `modelMap`.

#### 4. Remove old functions
**File**: `internal/workflow/workflow.go`
**Changes**: Remove `transformCommandFrontmatter`, `transformAgentFrontmatter`. Update `Install()` backward-compat wrapper to call new function.

#### 5. Tests
**File**: `internal/workflow/workflow_test.go`
**Changes**:
- Update `InstallTo` tests → `InstallSkills` tests covering all three targets
- Test `injectFrontmatter` with various field combinations
- Test that canonical installs have no tool-specific fields (AS-10, TC-4)
- Test content parity between canonical and enriched copies (AS-11, TC-5)
- Update prompt structure validation tests for new `skills/*/SKILL.md` paths
- Validate all 14 skills present (AS-12)
- Validate name matches directory and regex (AS-7, AS-8, TC-3)

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/workflow/...` passes (18/18)
- [x] `go build ./...` succeeds

### Commit:
- [ ] Stage: `internal/workflow/workflow.go`, `internal/workflow/workflow_test.go`
- [ ] Message: `feat: dual-directory Agent Skills install with frontmatter injection`

---

## Phase 3: Update init and update commands

### Overview
Wire new install logic into CLI commands. Add `agents-only` target. Stop creating `commands/` and `agents/` subdirs. Install to `.agents/skills/` for all targets.

### Tasks:

#### 1. Update `resolveTargetConfig`
**File**: `cmd/rpi/init_cmd.go`
**Changes**:
- `claude` target: subdirs → `["skills", "hooks"]` (remove `commands`, `agents`) — AS-14
- `opencode` target: subdirs → `["skills", "hooks"]`
- Add `agents-only` target: `toolDir=""`, no subdirs, no rules file, `TargetAgentsOnly`
- Update `--target` flag help text to include `agents-only`

#### 2. Update `runInit` for dual-directory install
**File**: `cmd/rpi/init_cmd.go`
**Changes**:
- Create `.agents/skills/` directory for all targets
- Call `workflow.InstallSkills(agentsSkillsDir, toolDirPath, target, false)`
- For `agents-only`: skip tool dir creation, rules file, MCP config, settings
- Don't add `.agents/` to `.gitignore` (AS spec: skills should be shared)
- Update log messages to reflect new structure

#### 3. Update `runUpdate` for dual-directory install
**File**: `cmd/rpi/update_cmd.go`
**Changes**:
- `detectTarget` also checks for `.agents/` (for `agents-only` projects)
- Ensure `.agents/skills/` exists
- Call `workflow.InstallSkills` with both directories
- Don't delete or modify existing `.claude/commands/` (AS-13)

#### 4. Tests
**Files**: `cmd/rpi/init_cmd_test.go`, `cmd/rpi/update_cmd_test.go`
**Changes**:
- **TC-1**: Fresh claude init → `.agents/skills/` has 14 dirs + `.claude/skills/` has 14 dirs + no `.claude/commands/` or `.claude/agents/`
- **TC-2**: Fresh agents-only init → `.agents/skills/` has 14 dirs + no `.claude/` or `.opencode/`
- **TC-6**: Update with existing commands dir → `.agents/skills/` created + `.claude/commands/` untouched + `.claude/skills/` created
- Update existing tests for removed `commands/`/`agents/` subdirs
- Test `--target agents-only` flag
- Test `.agents/` not in `.gitignore`

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/...` passes
- [x] `go test ./internal/...` passes
- [x] `go build ./...` succeeds

### Commit:
- [ ] Stage: `cmd/rpi/init_cmd.go`, `cmd/rpi/update_cmd.go`, `cmd/rpi/init_cmd_test.go`, `cmd/rpi/update_cmd_test.go`
- [ ] Message: `feat: agents-only target and dual-directory skill install in init/update`

---

## Phase 4: End-to-end validation and spec transition

### Overview
Run full test suite, verify all spec behaviors, transition artifacts.

### Tasks:

#### 1. Full test suite
Run `go test ./...` and fix any remaining failures.

#### 2. Verify spec coverage
Cross-check each AS-* behavior ID against test coverage:
- AS-1 through AS-5: Installation targets (TC-1, TC-2, TC-6)
- AS-6 through AS-10: Format compliance (TC-3, TC-4)
- AS-11 through AS-12: Content integrity (TC-5)
- AS-13 through AS-14: Backward compatibility (TC-1, TC-6)

#### 3. Transition artifacts
- Design → `complete`
- Spec → `active`

### Success Criteria:

#### Automated Verification:
- [ ] `go test ./...` passes (all packages)
- [ ] `go vet ./...` clean

#### Manual Verification:
- [ ] Every AS-* behavior ID has at least one covering test

### Commit:
- [ ] Stage: any remaining fixes
- [ ] Message: `test: complete Agent Skills spec coverage`

---

## References
- Design: .rpi/designs/2026-03-22-agent-skills-compatibility.md
- Spec: .rpi/specs/agent-skills-compatibility.md
