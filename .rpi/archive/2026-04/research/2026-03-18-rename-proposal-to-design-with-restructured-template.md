---
archived_date: "2026-04-02"
branch: main
date: 2026-03-18T23:59:03+01:00
git_commit: 5de3542
repository: ai-agent-research-plan-implement-flow
researcher: Claude
status: archived
tags:
    - research
    - workflow
    - artifacts
    - design
    - template
topic: rename proposal to design with restructured template
---

# Research: Rename Proposal to Design with Restructured Template

## Research Question

Should the "proposal" artifact be renamed to "design" with a restructured template that emphasizes components, interfaces, and data flow — and if so, what does that look like concretely?

## Problem Statement

The current "proposal" artifact does double duty — investigation/reasoning AND technical design. The architecture lives inside "Design Decisions → How it works" blocks, so understanding how the system is built requires reading through multiple decision records and mentally reconstructing the architecture. The name "proposal" emphasizes persuasion ("I'm proposing we do X") rather than structure ("here's how it's built"). A "design" artifact would put architecture front and center.

## Summary

Rename "proposal" to "design" and restructure the template to emphasize architecture. The spec remains a separate living document — it's a per-module contract that accumulates behaviors across multiple design cycles, while designs are per-change and get archived. Decisions are integrated into component/interface descriptions rather than being standalone structured records.

## Detailed Findings

### Current proposal template sections

```
Summary → Investigation Findings → Constraints & Requirements →
Design Decisions (Chosen/Alternatives/Rationale/How it works) →
Architecture (optional) → File Structure →
Risks & Mitigations → What This Does NOT Cover → Open Questions → References
```

Key observation: "Design Decisions" carries the actual architecture in "How it works" blocks. The "Architecture" section is optional and when present is a simple diagram. To understand how the system is built, you read through 4-5 decision records.

### Proposed design template sections

```
Summary → Context → Constraints →
Components (with integrated decisions) →
Interfaces → Data Flow →
File Structure → Risks → Out of Scope → References
```

Changes from proposal:
- **"Investigation Findings" → "Context"** — shorter, focused on why, not a full investigation dump
- **"Design Decisions" removed as standalone section** — decisions integrated into component/interface descriptions (Google Design Docs / RFC style)
- **New: "Components"** — what pieces exist/are being built, their responsibilities
- **New: "Interfaces"** — how components connect, APIs, data contracts
- **New: "Data Flow"** — runtime path, what happens when X is triggered
- **"Architecture" removed** — absorbed into Components + Interfaces + Data Flow
- **"What This Does NOT Cover" → "Out of Scope"** — simpler name
- **"Open Questions" removed** — questions should be resolved before the design is finalized

### Scaling to design size

Components, Interfaces, and Data Flow scale to complexity:
- **Small change**: may only need Components
- **Medium feature**: Components + Interfaces
- **Complex feature**: all three sections

### Spec stays separate

The spec is a per-module living contract that persists across design cycles:
- A first design creates the spec
- A later design updates it with new/changed behaviors
- The spec is the authority on "what does this module do right now"

This is a deliberate choice — the user wants specs to accumulate behaviors across multiple design cycles for longer projects. Merging spec into design would lose this, because designs get archived.

### Decision format: integrated over structured

Industry approaches:
1. **Integrated** (Google Design Docs, RFCs): Decisions woven into the design narrative — describe the component, mention alternatives, explain the choice inline
2. **Structured records** (ADRs): Each decision is a standalone record with Chosen/Alternatives/Rationale

Chosen: **Integrated only**. Decisions live inside the component/interface descriptions where they're contextually relevant. No separate "Decisions" section.

Example:
```markdown
## Components

### MCP Server
Exposes all RPI operations as typed MCP tools over stdio.
Uses flat tool granularity (one tool per CLI command) rather than
coarse grouping with action discriminators, because an action
parameter reintroduces the guessing problem this solves.
```

### Codebase touch points

**Go code (rename "proposal" → "design"):**
- `internal/template/render.go:29` — `Proposal string` → `Design string` on `RenderContext`
- `internal/template/render.go:38` — `"propose": "Proposal"` → `"design": "Design"` in typeLabels
- `internal/template/render.go:73` — `case "research", "propose":` → include `"design"`
- `cmd/rpi/scaffold.go` — `proposalFlag` → `designFlag`, `typeDirs["propose"]` → `typeDirs["design"]`, `"proposals"` → `"designs"`
- `cmd/rpi/serve.go` — `scaffoldInput.Proposal` → `Design`, `scanInput.Proposal` → `Design`
- `internal/chain/resolve.go:15` — `"proposal"` → `"design"` in `linkFields`
- `internal/chain/resolve.go:209-210` — `"proposals"` → `"designs"` in `inferType`
- `internal/scanner/scan.go:112,244-245` — `"proposal"` → `"design"` in field matching and type inference
- All corresponding test files

**Templates (rename + restructure):**
- `propose.tmpl` → `design.tmpl` (renamed + new section structure)
- `plan.tmpl` — `proposal:` → `design:` frontmatter field
- `spec.tmpl` — `updated_by` comment
- `verify-report.tmpl` — `proposal:` → `design:` frontmatter
- `CLAUDE.md.template` / `AGENTS.md.template` — directory descriptions
- `PIPELINE.md.template` — flow diagrams, stage descriptions

**Command files (rename + reframe):**
- `rpi-propose.md` → `rpi-design.md` (renamed + restructured)
- `rpi-plan.md` — references to "proposal" → "design"
- `rpi-implement.md` — references to "proposal" → "design"
- `rpi-verify.md` — references to "proposal" → "design"
- `rpi-archive.md` — references to "proposal" → "design"

**Directory rename:**
- `.rpi/proposals/` → `.rpi/designs/`

## Assessment

This is a meaningful change, not just a rename. The restructured template (Components, Interfaces, Data Flow with integrated decisions) changes how the agent thinks about and structures design artifacts — guiding it toward architecture documentation rather than decision records that happen to contain architecture. The spec staying separate preserves the living-contract model for longer projects.

The scope is significant — touches Go code, templates, all command files, pipeline docs, and directory naming. Should be combined with the stage transition hardening work since both restructure the command files.

## Suggested Next Steps

- Propose this change via `/rpi-propose`
- Consider combining with the transition hardening fix since both touch command files
- The Go code changes are mechanical (rename); the template restructure and command rewrite are the substantive work

## Decisions

- **Integrated decisions over structured records** — decisions woven into component descriptions, no standalone Decisions section
- **Spec stays separate** — living per-module contract, accumulates across design cycles
- **Sections scale to design size** — Components always present; Interfaces and Data Flow included as complexity warrants
