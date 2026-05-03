---
domain: agent-skills-compatibility
feature: agent-skills
last_updated: 2026-05-01T00:00:00+02:00
updated_by: .rpi/plans/2026-05-01-bundle-grill-me-skill-from-mattpocock-skills.md
---

# Agent Skills Compatibility

## Purpose

Ensure all RPI workflow files are installable as Agent Skills-compliant SKILL.md files, making them discoverable by any tool that implements the Agent Skills standard across multiple targets (Claude, OpenCode, agents-only).

## Scenarios

### Init installs skills for Claude target
Given an empty project directory
When the user runs `rpi init` (default Claude target)
Then `.claude/skills/` contains 11 skill subdirectories with SKILL.md files

### Init installs skills for agents-only target
Given an empty project directory
When the user runs `rpi init --target agents-only`
Then `.agents/skills/` contains 11 skill subdirectories and no `.claude/` or `.opencode/` directory exists

### Bundled third-party skills include attribution
Given a skill bundled from a third-party source (e.g. `grill-me` from mattpocock/skills)
When the user runs `rpi init` for any target
Then the skill's deployed directory contains both `SKILL.md` and a `LICENSE` file with the upstream copyright and permission notice

### Skills conform to Agent Skills format
Given all embedded SKILL.md files
When parsing their frontmatter
Then every file has `name` and `description` fields, the name matches its parent directory, and the name matches the naming regex

### Skills carry no tool-specific frontmatter
Given any deployed SKILL.md file (across all targets)
When checking its frontmatter
Then it contains only `name` and `description` and never `model`, `disable-model-invocation`, `tools`, `allowed-tools`, or `context`

### Skill content is identical across targets
Given a skill installed for any target
When comparing the markdown content
Then `SKILL.md` is byte-identical to the canonical embedded source

### Init preserves existing commands directory
Given a project with existing `.claude/commands/` files
When the user runs `rpi update`
Then `.claude/commands/` is left untouched and `.claude/skills/` is created or updated

## Constraints
- Follow Agent Skills naming: lowercase, hyphens, no consecutive hyphens, ≤64 chars
- Support all three targets: claude, opencode, agents-only
- Do not overwrite existing files without `--force`
- All 10 first-party pipeline skills must be present: rpi-research, rpi-propose, rpi-plan, rpi-implement, rpi-verify, rpi-diagnose, rpi-explain, rpi-commit, rpi-archive, rpi-spec-sync
- One bundled third-party skill ships alongside the pipeline skills: `grill-me` (sourced from https://github.com/mattpocock/skills under MIT). Its upstream `LICENSE` is deployed beside its `SKILL.md` so the permission notice travels with each copy.
- Sibling files in any skill source directory (e.g. `LICENSE`, `NOTICE`) are deployed verbatim alongside `SKILL.md`

## Out of Scope
- MCP server changes
- Prompt content rewrites
- New tool targets beyond claude/opencode
- Per-skill hooks or agent designations
- Plugin packaging
