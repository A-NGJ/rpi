---
date: 2026-03-23
topic: "agent-skills-compatibility"
tags: [review]
design: .rpi/designs/2026-03-22-agent-skills-compatibility.md
spec: .rpi/specs/agent-skills-compatibility.md
---

# Verification Report: Agent Skills Compatibility

## Summary

Verification of the Agent Skills compatibility refactoring (spec AS, commits b03b7ad..9062899). The implementation correctly converts all workflow files to Agent Skills format with per-target skill installation and frontmatter injection. One blocker found: deleted `.claude/templates/` files break the scaffold command and 9 tests.

## Completeness

### Spec Behaviors

| ID | Behavior | Status | Evidence |
|----|----------|--------|----------|
| AS-1 | `rpi init` creates skill dirs (9 total) with SKILL.md | PASS | `init_cmd.go:210` calls `InstallSkills`; `TestInitInstallsSkills` verifies 9 skills |
| AS-2 | `--target claude` creates `.claude/skills/` with overrides | PASS | `workflow.go:71-74` injects overrides for claude; `TestInstallSkills_Claude` verifies model fields |
| AS-3 | `--target opencode` creates `.opencode/skills/` with overrides | PASS | `TestInitOpenCode` verifies 9 dirs; `TestInstallSkills_OpenCode` verifies model translation |
| AS-4 | `--target agents-only` creates only `.agents/skills/` | PASS | `init_cmd.go:139-143`; `TestInitAgentsOnly` verifies no `.claude/` or `.opencode/` |
| AS-5 | `rpi update` updates skills in active directory | PASS | `update_cmd.go:110-123`; `TestUpdateInstallsSkills` (in update_cmd_test.go) |
| AS-6 | Every SKILL.md has `name` and `description` | PASS | `TestSkillNameMatchesDir` + `TestSkillDescriptionValid` |
| AS-7 | `name` matches parent directory | PASS | `TestSkillNameMatchesDir` |
| AS-8 | `name` matches regex, <=64 chars | PASS | `TestSkillNameMatchesDir` with regex `^[a-z][a-z0-9]*(-[a-z0-9]+)*$` |
| AS-9 | `description` 1-1024 chars with keywords | PASS | `TestSkillDescriptionValid` |
| AS-10 | Canonical SKILL.md has no tool-specific fields | PASS | `TestCanonicalSkillsHaveNoToolFields` |
| AS-11 | Body identical between canonical and tool-specific | PASS | `TestInstallSkills_ContentParity` |
| AS-12 | All 9 pipeline skills present | PASS | `TestAllSkillsPresent` — 9 embedded dirs verified |
| AS-13 | Existing `.claude/commands/` not deleted | PASS | Init/update never touch `commands/` directory |
| AS-14 | No `.claude/commands/` or `.claude/agents/` created | PASS | `TestInitCreatesAllDirs` explicitly checks absence |

### Files Changed

All 22 changed files are accounted for by the spec scope:
- `workflow.go` / `workflow_test.go` — core InstallSkills + tests
- `init_cmd.go` / `init_cmd_test.go` — single-directory install per target
- `update_cmd.go` / `update_cmd_test.go` — update path
- `internal/workflow/assets/skills/*/SKILL.md` — 9 pipeline skills (renamed from flat files)
- Removed: 5 non-pipeline skills (codebase-analyzer agent, analyze-thoughts, find-patterns, locate-codebase, locate-thoughts)
- Templates and docs updated

### TODO/FIXME/HACK Markers

7 TODO markers found — all are template placeholders (e.g., `<!-- TODO: Add brief project description -->`) intended to be filled by users. Not actionable.

## Correctness

### Blocker: Deleted `.claude/templates/` breaks scaffold command

**Files**: `.claude/templates/plan.tmpl`, `.claude/templates/verify-report.tmpl`
**Status**: Deleted from working tree (shown in `git status` as `D`)

The `rpi scaffold` command and `rpi serve` MCP server both default `--templates-dir` to `.claude/templates/` (`main.go:57`, `serve.go:395`). With these files deleted, 9 scaffold-related tests fail:
- `TestScaffoldPlanStdout`
- `TestScaffoldPlanWrite`
- `TestScaffoldPlanWriteOverwriteProtection`
- `TestScaffoldPlanWriteForce`
- `TestScaffoldAllTypes` (5 subtests)
- `TestScaffoldCustomTemplatesDir`
- `TestScaffoldPlanWithTicket`
- `TestScaffoldPlanWithoutTicket`
- `TestScaffoldPlanWithSpec`

**Root cause**: The template files were deleted (likely during the skills directory restructuring) but the scaffold code still references them. Either the deletions are unintended, or the scaffold command needs to be updated to use embedded templates.

### Warning: `copyDirectory` / `copyDirRecursive` appear unused

**File**: `init_cmd.go:367-423`

These two functions (`copyDirectory`, `copyDirRecursive`) are no longer called after the refactor to `workflow.InstallSkills`. They were likely used for the old directory-copy approach. `TestCopyDirectory` still tests them but they serve no production purpose.

## Coherence

### Note: Single-directory install diverges from spec AS-1 wording

The spec says `rpi init` MUST create `.agents/skills/` containing skills. The implementation installs to the tool-specific directory only (e.g., `.claude/skills/`), not `.agents/skills/` for claude/opencode targets. This is consistent with the latest commit message ("single-directory skill install per target") and appears to be an intentional simplification. The spec may need updating to reflect this decision — currently AS-1 through AS-3 imply `.agents/skills/` is always created.

### Note: No `.agents/` gitignore entry for agents-only

For `agents-only` target, `.agents/` is correctly NOT added to `.gitignore` (per spec "Must Not" constraint). Verified in `TestInitAgentsOnlyNotInGitignore`.

## Findings Summary

| Severity | Count | Description |
|----------|-------|-------------|
| Blocker  | 1     | Deleted `.claude/templates/` breaks scaffold command (9 test failures) |
| Warning  | 1     | Unused `copyDirectory`/`copyDirRecursive` functions in init_cmd.go |
| Note     | 1     | Spec AS-1/AS-2/AS-3 wording implies dual-directory but implementation is single-directory |

## Overall Status: WARN

All 14 Agent Skills spec behaviors (AS-1 through AS-14) pass. The blocker is outside the AS spec scope but affects the same codebase — the deleted template files break the scaffold/MCP functionality. The `internal/workflow` tests all pass (cached).

## Report Path

`.rpi/reviews/2026-03-23-agent-skills-compatibility.md`
