---
date: 2026-03-19T09:50:32+01:00
related_research: .rpi/research/2026-03-18-rename-proposal-to-design-with-restructured-template.md
status: complete
tags:
    - design
    - workflow
    - artifacts
    - template
topic: rename-proposal-to-design
---

# Design: Rename Proposal Artifact to Design

## Summary

Rename the "proposal" artifact type to "design" and restructure its template to emphasize architecture (Components with integrated decisions) rather than decision records. The `/rpi-propose` command stays — it's the action; the artifact it produces becomes a "design". Specs remain separate living documents.

## Context

The current "proposal" artifact does double duty: investigation/reasoning AND technical design. Architecture lives buried inside "Design Decisions → How it works" blocks, forcing readers to mentally reconstruct the system from multiple decision records. A "design" artifact puts architecture front and center.

The `/rpi-propose` command name captures the action (investigating, weighing trade-offs, proposing a direction). The artifact it produces — "design" — captures the output (how the system is built). This decoupling is intentional: the command name shouldn't change because "propose" accurately describes what the user is doing, while "design" accurately describes what they get.

## Constraints

- The flow hasn't been released yet, so backward compatibility is not a concern
- `.rpi/` is gitignored — no migration needed for existing artifacts
- Spec stays separate: it's a per-module living contract that accumulates across design cycles, while designs are per-change and get archived
- Template sections must scale: small changes need only Components; complex features add Interfaces and Data Flow via the command prompt, not the template

## Components

### Design template (`design.tmpl`)

Replaces `propose.tmpl` with restructured sections:

```
Summary → Context → Constraints → Components → File Structure → Risks → Out of Scope → References
```

Components is the core section where architecture lives. Each component describes its responsibility with decisions integrated inline rather than as structured records:

```markdown
## Components

### MCP Server
Exposes all RPI operations as typed MCP tools over stdio.
Uses flat tool granularity (one tool per CLI command) rather than
coarse grouping with action discriminators, because an action
parameter reintroduces the guessing problem this solves.
```

The template does NOT include Interfaces or Data Flow — those are added by the command prompt when complexity warrants.

Section changes from proposal template:
- "Investigation Findings" → "Context" (shorter, focused on why)
- "Design Decisions" removed — integrated into Components
- "Architecture" removed — absorbed into Components
- "What This Does NOT Cover" → "Out of Scope"
- "Open Questions" removed — must be resolved before finalizing

### Command file (`rpi-propose.md`)

Stays as `rpi-propose.md`, invoked as `/rpi-propose`. Updated to:
- Scaffold `type=design` instead of `type=propose`
- Reference "design" for the artifact throughout
- Quick mode: scaffolds design with Summary, Context, Constraints, Components, References
- Full mode: for medium/complex features, the command adds Interfaces and/or Data Flow sections after scaffolding

### Scaffold type mapping

The `rpi scaffold` command gets a new type `design` replacing `propose`:
- `rpi scaffold design --topic "foo"` → `.rpi/designs/YYYY-MM-DD-foo.md`
- The `propose` type is removed entirely (no aliases, no backward compat)

### Frontmatter field rename

All artifacts that reference a proposal via frontmatter change the field name:
- `plan.tmpl`: `proposal:` → `design:`
- `verify-report.tmpl`: `proposal:` → `design:`
- Chain resolution: `linkFields` includes `"design"` instead of `"proposal"`
- Scanner: filters on `design` field instead of `proposal`

### Go code changes

Mechanical renames across the codebase:

**Struct fields and variables:**
- `RenderContext.Proposal` → `RenderContext.Design` (`internal/template/render.go:29`)
- `proposalFlag` → `designFlag` (`cmd/rpi/scaffold.go:16`)
- `scaffoldInput.Proposal` → `scaffoldInput.Design` (`cmd/rpi/serve.go`)
- `scanInput.Proposal` → `scanInput.Design` (`cmd/rpi/serve.go`)
- `scanner.Filters.Proposal` → `scanner.Filters.Design` (`internal/scanner/scan.go:25`)

**Type mappings:**
- `typeDirs["propose"]` → `typeDirs["design"]` mapping to `"designs"` (`cmd/rpi/scaffold.go:27`)
- `typeLabels["propose"]` → `typeLabels["design"]` mapping to `"Design"` (`internal/template/render.go:38`)

**String constants:**
- `linkFields`: `"proposal"` → `"design"` (`internal/chain/resolve.go:15`)
- Type inference: `"proposals"` → `"designs"` (`internal/chain/resolve.go:209-210`, `internal/scanner/scan.go:244-245`)
- Valid types in CLI help: `"propose"` → `"design"` (`cmd/rpi/scaffold.go:34`, `cmd/rpi/scan.go:28`)
- Init directories: `"proposals"` → `"designs"` (`cmd/rpi/init_cmd.go`)

### Documentation templates

- `CLAUDE.md.template` / `AGENTS.md.template`: `.rpi/proposals/` → `.rpi/designs/`, description text updated
- `PIPELINE.md.template`: artifact references updated, pipeline stage description stays "Propose" (it's the action)
- `CLAUDE.md` (project root): updated to match

## File Structure

**Renamed files:**
- `internal/workflow/assets/templates/propose.tmpl` → `design.tmpl`
- `.claude/templates/propose.tmpl` → `design.tmpl`
- `.opencode/templates/propose.tmpl` → `design.tmpl`

**Modified files (Go code):**
- `internal/template/render.go`
- `internal/template/render_test.go`
- `cmd/rpi/scaffold.go`
- `cmd/rpi/scaffold_test.go`
- `cmd/rpi/serve.go`
- `cmd/rpi/scan.go`
- `cmd/rpi/init_cmd.go`
- `cmd/rpi/init_cmd_test.go`
- `cmd/rpi/main.go`
- `cmd/rpi/chain.go`
- `cmd/rpi/archive.go`
- `cmd/rpi/extract.go`
- `internal/chain/resolve.go`
- `internal/chain/resolve_test.go`
- `internal/scanner/scan.go`
- `internal/scanner/scan_test.go`

**Modified files (templates):**
- `plan.tmpl` (all 3 locations)
- `verify-report.tmpl` (all 3 locations)
- `CLAUDE.md.template` (all 3 locations)
- `AGENTS.md.template` (all 3 locations)
- `PIPELINE.md.template` (all 3 locations)

**Modified files (commands):**
- `rpi-propose.md` (all 3 locations) — content rewritten, filename stays
- `rpi-plan.md`, `rpi-implement.md`, `rpi-verify.md`, `rpi-archive.md` (all 3 locations) — reference updates

**Modified files (project root):**
- `CLAUDE.md`

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Large scope (~45 files) | Missed references | Grep for `propos` after implementation to catch stragglers |
| Template restructure changes agent behavior | Designs may be less thorough initially | Command prompt guides section usage; iterate on prompt |

## Out of Scope

- Renaming the `/rpi-propose` command itself
- Changes to spec template structure
- Adding Interfaces/Data Flow to the template (command prompt handles this)
- Migration tooling for existing `.rpi/proposals/` directories

## References

- Research: .rpi/research/2026-03-18-rename-proposal-to-design-with-restructured-template.md
