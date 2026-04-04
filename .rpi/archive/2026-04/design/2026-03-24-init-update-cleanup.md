---
archived_date: "2026-04-02"
date: 2026-03-24T00:47:14+01:00
spec: .rpi/specs/init-update.md
status: archived
tags:
    - design
    - refactor
topic: init-update-cleanup
---

# Design: init-update-cleanup

## Summary

Extract shared project-sync logic from `init` and `update` into a common `syncProject()` function, and fix the rules file overwrite inconsistency so `--force` governs all managed files uniformly.

## Context

`rpi init` and `rpi update` share ~80% of their logic (create dirs, install skills/templates, generate rules file, build index) but implement it independently with copy-pasted code. This creates two problems:

1. **Inconsistency**: `update` always overwrites the rules file (CLAUDE.md/AGENTS.md) regardless of `--force`, but respects `--force` for skills and templates. A user who customized their rules file loses changes on every `update`.
2. **Maintenance burden**: The `.rpi/` subdir list, skill install logic, and index build logic are duplicated. Adding a new subdir or changing install behavior requires editing both files.

The commands themselves are justified as separate — `init` for first-time setup, `update` for syncing to latest — following standard CLI conventions (git, terraform, npm).

## Constraints

- Must not change the public CLI interface (command names, flag names, flag defaults)
- Must not break any existing tests (33 tests across both files)
- Must preserve `init`'s one-time-only behaviors: guard against re-init, `.gitignore` management, MCP registration
- Must preserve `update`'s auto-detection of target from existing directories

## Components

### `syncProject()` — shared core function

A new unexported function in a shared location (either `init_cmd.go` or a new `sync.go` file) that both commands call:

```go
type syncOptions struct {
    targetDir    string
    cfg          targetConfig
    force        bool    // governs skills, templates, AND rules file
    skipRules    bool    // --no-claude-md
    w            io.Writer
}

func syncProject(opts syncOptions) error
```

This function handles:
1. Ensure `.rpi/` subdirs exist (create only if missing)
2. Ensure tool subdirs exist (create only if missing)
3. Install skills to target skills dir (respects `force`)
4. Install scaffold templates (respects `force`)
5. Update rules file (respects `force` — **the fix**)
6. Ensure settings.json has MCP permissions (Claude only)
7. Rebuild codebase index

Both `init` and `update` call `syncProject()` after their own pre-flight logic.

**Alternative considered**: putting `syncProject` in `internal/workflow/`. Rejected because it depends on CLI-layer concerns (logging, cobra output writers) and would create a circular dependency with the index package.

A new file `cmd/rpi/sync.go` is cleanest — it keeps the shared logic co-located with the commands that use it, and avoids bloating either command file.

### `runInit` — simplified

After extracting shared logic, `runInit` becomes:
1. Resolve target from `--target` flag
2. Guard: fail if tool dir (or `.agents/`) already exists
3. Create tool dir + subdirs (first-time creation, not idempotent check)
4. Call `syncProject()` with `force=false` (files are new, so everything writes)
5. Manage `.gitignore` entries
6. Configure MCP server (Claude only, unless `--no-mcp`)

### `runUpdate` — simplified

After extracting shared logic, `runUpdate` becomes:
1. Guard: fail if `.rpi/` doesn't exist
2. Auto-detect target from existing directories
3. Call `syncProject()` with `force=updateForce`

### Rules file behavior change

Current: `update` always overwrites the rules file.
New: `update` only overwrites the rules file when `--force` is passed, consistent with skills/templates.

This is a **behavioral change** but aligns with user expectations. The `--force` flag description already says "Overwrite existing workflow files" — rules files are workflow files.

## File Structure

| File | Change |
|---|---|
| `cmd/rpi/sync.go` | **New** — `syncProject()` function, `syncOptions` type |
| `cmd/rpi/init_cmd.go` | **Modified** — `runInit` calls `syncProject()`, remove duplicated logic. Keep `ensureGitignoreEntry`, `configureMCP`, `configureSettings`, logging helpers, `resolveTargetConfig`, `targetConfig` |
| `cmd/rpi/update_cmd.go` | **Modified** — `runUpdate` calls `syncProject()`, remove duplicated logic. Keep `detectTarget` |
| `cmd/rpi/update_cmd_test.go` | **Modified** — `TestUpdateUpdatesRulesFile` updated to expect rules file NOT overwritten without `--force`. Add test for `--force` overwriting rules file |

## Risks

1. **Test breakage from rules file behavior change**: `TestUpdateUpdatesRulesFile` currently asserts that `update` overwrites CLAUDE.md. This test must be updated to match the new behavior. Low risk — the test is explicitly testing the behavior we're changing.
2. **Shared state in logging helpers**: `logSuccess`/`logWarning`/`logInfo` are defined in `init_cmd.go`. Moving `syncProject` to `sync.go` means these helpers must be accessible from `sync.go`. Since they're in the same package, this works without changes.

## Out of Scope

- Merging `init` and `update` into a single command
- Changing flag names or defaults
- Adding new features to either command
- Updating the MCP spec (`mcp-init-and-commands.md`) — no MCP behavior changes
