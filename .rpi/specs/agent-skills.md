---
domain: agent-skills-compatibility
feature: agent-skills
last_updated: 2026-04-04T22:30:00+02:00
updated_by: .rpi/designs/2026-03-22-agent-skills-compatibility.md
---

# Agent Skills Compatibility

## Purpose

Ensure all RPI workflow files are installable as Agent Skills-compliant SKILL.md files, making them discoverable by any tool that implements the Agent Skills standard across multiple targets (Claude, OpenCode, agents-only).

## Scenarios

### Init installs skills for Claude target
Given an empty project directory
When the user runs `rpi init` (default Claude target)
Then `.claude/skills/` contains 9 skill subdirectories with SKILL.md files including Claude-specific frontmatter overrides

### Init installs skills for agents-only target
Given an empty project directory
When the user runs `rpi init --target agents-only`
Then `.agents/skills/` contains 9 skill subdirectories and no `.claude/` or `.opencode/` directory exists

### Skills conform to Agent Skills format
Given all embedded SKILL.md files
When parsing their frontmatter
Then every file has `name` and `description` fields, the name matches its parent directory, and the name matches the naming regex

### Canonical skills have no tool-specific fields
Given all canonical SKILL.md files in the embedded assets
When checking their frontmatter
Then none contain `model`, `disable-model-invocation`, or `tools` fields

### Skill content is identical across targets
Given a skill installed for both the canonical and a tool-specific target
When comparing the markdown body content
Then the body is identical between copies — only frontmatter differs

### Init preserves existing commands directory
Given a project with existing `.claude/commands/` files
When the user runs `rpi update`
Then `.claude/commands/` is left untouched and `.claude/skills/` is created or updated

## Constraints
- Follow Agent Skills naming: lowercase, hyphens, no consecutive hyphens, ≤64 chars
- Support all three targets: claude, opencode, agents-only
- Do not overwrite existing files without `--force`
- All 9 pipeline skills must be present: rpi-research, rpi-propose, rpi-plan, rpi-implement, rpi-verify, rpi-diagnose, rpi-explain, rpi-commit, rpi-archive

## Out of Scope
- MCP server changes
- Prompt content rewrites
- New tool targets beyond claude/opencode
- Agent Skills `allowed-tools` field
