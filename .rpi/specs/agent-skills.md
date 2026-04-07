---
domain: agent-skills-compatibility
feature: agent-skills
last_updated: 2026-04-07T16:00:00+02:00
updated_by: .rpi/designs/2026-04-07-skill-metadata-for-read-only-safety.md
---

# Agent Skills Compatibility

## Purpose

Ensure all RPI workflow files are installable as Agent Skills-compliant SKILL.md files, making them discoverable by any tool that implements the Agent Skills standard across multiple targets (Claude, OpenCode, agents-only).

## Scenarios

### Init installs skills for Claude target
Given an empty project directory
When the user runs `rpi init` (default Claude target)
Then `.claude/skills/` contains 10 skill subdirectories with SKILL.md files including Claude-specific frontmatter overrides

### Init installs skills for agents-only target
Given an empty project directory
When the user runs `rpi init --target agents-only`
Then `.agents/skills/` contains 10 skill subdirectories and no `.claude/` or `.opencode/` directory exists

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

### Read-only skills restricted from file modification
Given a project initialized for the Claude target
When inspecting the installed research, verify, and explain skills
Then their frontmatter includes an allowed-tools field that excludes Write, Edit, and NotebookEdit

### Research skill runs in isolated context
Given a project initialized for the Claude target
When inspecting the installed research skill
Then its frontmatter includes a context field set to fork

### Skill metadata not applied for agents-only target
Given a project initialized with the agents-only target
When inspecting the installed research, verify, and explain skills
Then their frontmatter does not contain allowed-tools or context fields

## Constraints
- Follow Agent Skills naming: lowercase, hyphens, no consecutive hyphens, ≤64 chars
- Support all three targets: claude, opencode, agents-only
- Do not overwrite existing files without `--force`
- All 10 pipeline skills must be present: rpi-research, rpi-propose, rpi-plan, rpi-implement, rpi-verify, rpi-diagnose, rpi-explain, rpi-commit, rpi-archive, rpi-spec-sync

## Out of Scope
- MCP server changes
- Prompt content rewrites
- New tool targets beyond claude/opencode
- Per-skill hooks or agent designations
- Plugin packaging
