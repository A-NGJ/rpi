---
domain: init-update-cleanup
feature: init-update
last_updated: 2026-04-04T22:30:00+02:00
updated_by: .rpi/designs/2026-03-24-init-update-cleanup.md
---

# Init/Update Cleanup

## Purpose

Ensure `rpi init` and `rpi update` share a single code path for project synchronization, and that `--force` uniformly governs overwriting of all managed files (skills, templates, and rules file).

## Scenarios

### Init and update share sync logic
Given the codebase
When inspecting the init and update command implementations
Then both call a shared sync function for installing skills, templates, rules file, and settings

### Update without --force preserves rules file
Given an initialized project with a customized CLAUDE.md
When `rpi update` runs without `--force`
Then the CLAUDE.md content remains unchanged

### Update with --force overwrites rules file
Given an initialized project with a customized CLAUDE.md
When `rpi update --force` runs
Then CLAUDE.md is replaced with the latest template version

### Init creates fresh project with rules file
Given an empty directory
When `rpi init` runs
Then a `.rpi/` directory is created with subdirectories, CLAUDE.md is written, and the MCP server is registered

### Init rejects existing tool directory
Given a directory with `.claude/` already present
When `rpi init` runs
Then it fails with an error indicating the tool directory already exists

### Update auto-detects target from existing directories
Given a project initialized with `--target opencode`
When `rpi update` runs without specifying a target
Then it detects the OpenCode target and operates on `.opencode/`

## Constraints
- Init and update remain separate Cobra commands
- Rules file overwrite governed by `--force`, same as skills/templates
- Do not change CLI interface (command names, flag names, flag defaults)
- Do not introduce new dependencies or packages

## Out of Scope
- Merging init/update into one command
- New flags or features
- Changes to skill content or template content
