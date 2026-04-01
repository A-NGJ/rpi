---
domain: AS
id: AS
last_updated: 2026-03-22T23:19:52+01:00
status: active
updated_by: .rpi/designs/2026-03-22-agent-skills-compatibility.md
---

# Agent Skills Compatibility

## Purpose

Ensure all RPI workflow files are installable as Agent Skills-compliant SKILL.md files in `.agents/skills/`, making them discoverable by any tool that implements the Agent Skills standard.

## Behavior

### Installation
- **AS-1**: `rpi init` MUST install one subdirectory per skill (9 total), each with a `SKILL.md`, to the target's skills directory
- **AS-2**: `rpi init --target claude` MUST create `.claude/skills/` with Claude-specific frontmatter overrides
- **AS-3**: `rpi init --target opencode` MUST create `.opencode/skills/` with OpenCode-specific frontmatter overrides
- **AS-4**: `rpi init --target agents-only` MUST create `.agents/skills/` — no tool-specific directory, no MCP config
- **AS-5**: `rpi update` MUST update skills in the active target's skills directory

### Agent Skills Format Compliance
- **AS-6**: Every installed `SKILL.md` MUST have `name` and `description` in YAML frontmatter
- **AS-7**: The `name` field MUST match its parent directory name exactly
- **AS-8**: The `name` field MUST match `^[a-z][a-z0-9]*(-[a-z0-9]+)*$` and be ≤64 characters
- **AS-9**: The `description` field MUST be 1-1024 characters and include activation keywords
- **AS-10**: Canonical (embedded) SKILL.md files MUST NOT contain tool-specific fields (`model`, `disable-model-invocation`, `tools`)

### Content Integrity
- **AS-11**: Prompt body (goal, invariants, principles sections) MUST be identical between `.agents/skills/` and tool-specific copies
- **AS-12**: All 9 pipeline skills MUST be present: rpi-research, rpi-propose, rpi-plan, rpi-implement, rpi-verify, rpi-diagnose, rpi-explain, rpi-commit, rpi-archive

### Backward Compatibility
- **AS-13**: Existing `.claude/commands/` files MUST NOT be deleted by `rpi init` or `rpi update`
- **AS-14**: `rpi init` MUST NOT create `.claude/commands/` or `.claude/agents/` directories

## Constraints

### Must
- Follow Agent Skills naming: lowercase, hyphens, no consecutive hyphens, ≤64 chars
- Preserve all existing prompt content unchanged
- Support all three targets: claude, opencode, agents-only

### Must Not
- Add `.agents/skills/` to `.gitignore` (skills should be shared)
- Overwrite existing files without `--force`

### Out of Scope
- MCP server changes
- Prompt content rewrites
- New tool targets beyond claude/opencode
- Agent Skills `allowed-tools` field

## Test Cases

### TC-1: Fresh claude init
- **Given** empty project directory **When** `rpi init` **Then** `.agents/skills/` has 9 dirs with valid SKILL.md AND `.claude/skills/` has 9 dirs with model fields AND no `.claude/commands/` or `.claude/agents/` exists

### TC-2: Fresh agents-only init
- **Given** empty project directory **When** `rpi init --target agents-only` **Then** `.agents/skills/` has 9 dirs AND no `.claude/` or `.opencode/` directory exists

### TC-3: Name validation
- **Given** all embedded SKILL.md files **When** parsing frontmatter **Then** every `name` matches parent dir AND matches naming regex AND ≤64 chars

### TC-4: No tool-specific fields in cross-tool skills
- **Given** all canonical `skills/*/SKILL.md` **When** parsing frontmatter **Then** none contain `model`, `disable-model-invocation`, or `tools`

### TC-5: Content parity
- **Given** a skill with both canonical and override versions **When** comparing body content **Then** markdown body is identical (only frontmatter differs)

### TC-6: Update with existing commands dir
- **Given** project with `.claude/commands/rpi-propose.md` **When** `rpi update` **Then** `.agents/skills/` is created AND `.claude/commands/` is untouched AND `.claude/skills/` is created
