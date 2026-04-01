---
archived_date: "2026-04-02"
date: 2026-03-22T23:19:47+01:00
status: archived
tags:
    - design
    - agent-skills
    - cross-tool
topic: Agent Skills Compatibility
---

# Design: Agent Skills Compatibility

## Summary

Restructure all RPI workflow files (9 commands + 4 skills + 1 agent) into Agent Skills-compliant format and install them to `.agents/skills/` — the cross-tool interoperability path adopted by 30+ agentic coding tools. Tool-specific copies (`.claude/skills/`, `.opencode/skills/`) provide enhanced UX where supported.

## Context

RPI currently installs workflow files to `.claude/commands/`, `.claude/skills/`, and `.claude/agents/` — all Claude Code-specific paths. OpenCode support exists via frontmatter transforms, but the files still land in `.opencode/`.

The Agent Skills standard (agentskills.io) defines a universal format: a directory with a `SKILL.md` containing YAML frontmatter (`name` + `description` required) and markdown instructions. The cross-tool path is `.agents/skills/`. This standard is adopted by Claude Code, Cursor, VS Code Copilot, OpenAI Codex, Gemini CLI, Goose, Amp, JetBrains Junie, Roo Code, Kiro, and 20+ more.

Key spec details:
- No "command" or "agent" concept — everything is a skill
- Progressive disclosure: name+description loaded at startup, full body on activation
- No MCP binding — skills and MCP are orthogonal
- No argument passing — skills are activated, not parameterized
- Cross-tool path: `.agents/skills/`; client-specific: `.<client>/skills/`

## Constraints

- Agent Skills `name` must match directory name, lowercase+hyphens only, ≤64 chars
- Agent Skills `description` must be non-empty, ≤1024 chars
- No tool-specific fields (`model`, `tools`) allowed in cross-tool SKILL.md
- Claude Code reads skills from both `.agents/skills/` and `.claude/skills/`
- Existing projects may have `.claude/commands/` — can't break them silently

## Components

### 1. Canonical Skills (`.agents/skills/`)

All 14 workflow files become Agent Skills-compliant directories:

| Current File | New Skill Directory |
|---|---|
| `commands/rpi-research.md` | `rpi-research/SKILL.md` |
| `commands/rpi-propose.md` | `rpi-propose/SKILL.md` |
| `commands/rpi-plan.md` | `rpi-plan/SKILL.md` |
| `commands/rpi-implement.md` | `rpi-implement/SKILL.md` |
| `commands/rpi-verify.md` | `rpi-verify/SKILL.md` |
| `commands/rpi-diagnose.md` | `rpi-diagnose/SKILL.md` |
| `commands/rpi-explain.md` | `rpi-explain/SKILL.md` |
| `commands/rpi-commit.md` | `rpi-commit/SKILL.md` |
| `commands/rpi-archive.md` | `rpi-archive/SKILL.md` |
| `skills/find-patterns/SKILL.md` | `find-patterns/SKILL.md` (no change) |
| `skills/analyze-thoughts/SKILL.md` | `analyze-thoughts/SKILL.md` (no change) |
| `skills/locate-thoughts/SKILL.md` | `locate-thoughts/SKILL.md` (no change) |
| `skills/locate-codebase/SKILL.md` | `locate-codebase/SKILL.md` (no change) |
| `agents/codebase-analyzer.md` | `codebase-analyzer/SKILL.md` |

Frontmatter: only `name` and `description`. Body: identical goal+invariants+principles content.

### 2. Tool-Specific Overrides

Claude Code and OpenCode copies add tool-specific frontmatter. These live in `.claude/skills/` and `.opencode/skills/` respectively.

Claude override additions: `model` (inherit/haiku/sonnet/opus), `disable-model-invocation`.
OpenCode override additions: `model` (full provider-qualified ID).

The override mechanism: if a tool-specific SKILL.md exists in `overrides/<target>/skills/`, it's installed to the tool directory. Otherwise, the canonical version is copied.

### 3. Embedded Assets Restructure

```
internal/workflow/assets/
├── skills/                    # Canonical Agent Skills (cross-tool)
│   ├── rpi-research/SKILL.md
│   ├── rpi-propose/SKILL.md
│   ├── ... (14 total)
│   └── codebase-analyzer/SKILL.md
└── overrides/
    ├── claude/skills/         # Claude-specific versions
    │   ├── rpi-commit/SKILL.md
    │   ├── rpi-archive/SKILL.md
    │   └── codebase-analyzer/SKILL.md
    └── opencode/skills/       # OpenCode-specific versions
        └── codebase-analyzer/SKILL.md
```

Most skills don't need overrides — only those with non-default model or special fields. The `workflow.InstallTo()` function merges canonical + overrides.

### 4. `rpi init` / `rpi update` Changes

New `--target` values: `claude` (default), `opencode`, `agents-only`.

All targets install to `.agents/skills/`. `claude` and `opencode` also install to their tool-specific directory. `agents-only` skips tool-specific setup (no MCP config, no tool directory).

The `.claude/commands/` and `.claude/agents/` directories are no longer created. Existing ones are left untouched for backward compatibility.

## File Structure

**New/modified Go files:**
- `internal/workflow/workflow.go` — new install logic for dual-directory + overrides
- `cmd/rpi/init_cmd.go` — add `agents-only` target, install to `.agents/skills/`
- `cmd/rpi/update_cmd.go` — same dual-install in update path

**New embedded assets:**
- `internal/workflow/assets/skills/*/SKILL.md` — 14 canonical skills
- `internal/workflow/assets/overrides/claude/skills/*/SKILL.md` — Claude overrides
- `internal/workflow/assets/overrides/opencode/skills/*/SKILL.md` — OpenCode overrides

**Removed embedded assets:**
- `internal/workflow/assets/commands/` — merged into skills
- `internal/workflow/assets/agents/` — merged into skills

## Risks

- **Other tools may not support slash-command activation of skills** — Mitigated by writing good descriptions that enable model-driven activation
- **Claude Code skill dedup** — If the same skill exists in both `.agents/skills/` and `.claude/skills/`, Claude Code may show duplicates. Need to test behavior. If so, skip `.agents/` install for claude target.
- **Breaking existing projects** — Old `.claude/commands/` files are left in place; `rpi update` doesn't delete them

## Out of Scope

- MCP server changes (already tool-agnostic)
- Prompt content rewrites (only format/frontmatter changes)
- Adding new tool targets beyond Claude Code and OpenCode
- Agent Skills `allowed-tools` field (experimental, not widely supported)

## References

- Agent Skills Specification: https://agentskills.io/specification
- Agent Skills example skills: https://github.com/anthropics/skills
