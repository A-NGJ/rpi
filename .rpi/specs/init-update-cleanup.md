---
domain: init-update-cleanup
id: IU
last_updated: 2026-03-24T00:47:14+01:00
status: active
updated_by: .rpi/designs/2026-03-24-init-update-cleanup.md
---

# Init/Update Cleanup

## Purpose

Ensure `rpi init` and `rpi update` share a single code path for project synchronization, and that `--force` uniformly governs overwriting of all managed files (skills, templates, and rules file).

## Behavior

### Shared sync logic
- **IU-1**: Both `rpi init` and `rpi update` use the same `syncProject()` function for: ensuring dirs, installing skills, installing templates, updating rules file, configuring settings.json, and rebuilding the index
- **IU-2**: The `.rpi/` subdirectory list is defined in exactly one place (not duplicated)

### Rules file respects --force
- **IU-3**: `rpi update` without `--force` does NOT overwrite an existing rules file (CLAUDE.md/AGENTS.md)
- **IU-4**: `rpi update --force` overwrites the rules file with the latest template
- **IU-5**: `rpi update --no-claude-md` skips the rules file regardless of `--force`
- **IU-6**: `rpi init` always writes the rules file (new project, file doesn't exist yet)

### Preserved init-only behaviors
- **IU-7**: `rpi init` fails with an error if the tool directory (`.claude/`, `.opencode/`, or `.agents/`) already exists
- **IU-8**: `rpi init` adds `.rpi/index.json` and the tool directory to `.gitignore`
- **IU-8a**: `rpi init --no-track` also adds `.rpi/` to `.gitignore`
- **IU-9**: `rpi init` registers the MCP server via `claude mcp add` (Claude target, unless `--no-mcp`)
- **IU-10**: `rpi init` accepts `--target` to select the AI tool

### Preserved update-only behaviors
- **IU-11**: `rpi update` fails if `.rpi/` does not exist
- **IU-12**: `rpi update` auto-detects the target from existing directories (`.claude/`, `.opencode/`, `.agents/`)
- **IU-13**: `rpi update` accepts `--force` to overwrite existing managed files

### No regressions
- **IU-14**: All existing `init` tests pass (adjusted for shared code path)
- **IU-15**: All existing `update` tests pass (adjusted for rules file behavior change)

## Constraints

### Must
- Extract shared logic into a single `syncProject()` function
- Keep `init` and `update` as separate cobra commands
- Rules file overwrite governed by `--force`, same as skills/templates

### Must Not
- Change CLI interface (command names, flag names, flag defaults)
- Change MCP registration behavior
- Change `.gitignore` behavior
- Introduce new dependencies or packages

### Out of Scope
- Merging init/update into one command
- New flags or features
- Changes to skill content or template content
- MCP spec updates

## Test Cases

### IU-1: Shared sync function exists
- **Given** the codebase **When** inspecting `cmd/rpi/sync.go` **Then** a `syncProject` function exists that is called by both `runInit` and `runUpdate`

### IU-2: Single subdir list
- **Given** the codebase **When** searching for the `.rpi/` subdirectory list **Then** it appears exactly once (in `sync.go`)

### IU-3: Update preserves rules file without --force
- **Given** an initialized project with a customized CLAUDE.md **When** `rpi update` runs without `--force` **Then** CLAUDE.md content is unchanged

### IU-4: Update --force overwrites rules file
- **Given** an initialized project with a customized CLAUDE.md **When** `rpi update --force` runs **Then** CLAUDE.md is replaced with the template version

### IU-5: Update --no-claude-md skips rules file
- **Given** an initialized project with a customized CLAUDE.md **When** `rpi update --no-claude-md --force` runs **Then** CLAUDE.md content is unchanged

### IU-6: Init writes rules file on fresh project
- **Given** an empty directory **When** `rpi init` runs **Then** CLAUDE.md is created with template content

### IU-7: Init guard rejects existing tool dir
- **Given** a directory with `.claude/` present **When** `rpi init` runs **Then** it fails with ".claude/ already exists"

### IU-12: Update auto-detects target
- **Given** a project initialized with `--target opencode` **When** `rpi update` runs **Then** it detects opencode and operates on `.opencode/`
