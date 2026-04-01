---
archived_date: "2026-04-02"
date: 2026-03-19T10:12:38+01:00
proposal: .rpi/proposals/2026-03-19-rename-proposal-to-design.md
spec: .rpi/specs/design-artifact.md
status: archived
tags:
    - plan
topic: rename-proposal-to-design
---

# rename-proposal-to-design — Implementation Plan

## Overview

Rename the "proposal" artifact type to "design" across Go code, templates, command files, and documentation. The `/rpi-propose` command stays — only the artifact it produces changes.

**Scope**: ~45 files modified, 3 files renamed (propose.tmpl → design.tmpl in 3 locations)

## Source Documents
- **Proposal**: .rpi/proposals/2026-03-19-rename-proposal-to-design.md
- **Spec**: .rpi/specs/design-artifact.md

## Phase 1: Go source code + tests

### Overview
Mechanical rename of "proposal/propose/proposals" → "design/designs" across all Go source files and their tests. After this phase, `go test ./...` passes and `rpi scaffold design` works.

**Spec behaviors**: DA-1, DA-3, DA-4, DA-5, DA-6, DA-7, DA-8, DA-9, DA-10, DA-11, DA-12

### Tasks:

#### 1. internal/template/render.go
- Line 29: `Proposal string` → `Design string` on RenderContext
- Line 38: `"propose": "Proposal"` → `"design": "Design"` in typeLabels
- Line 73: `case "research", "propose":` → `case "research", "design":` in GenerateFilename

#### 2. internal/template/render_test.go
- Line 49: `{"propose", ctx, ...}` → `{"design", ctx, ...}`
- Line 189: `Proposal: ".rpi/proposals/..."` → `Design: ".rpi/designs/..."`
- Line 201: `{"propose", []string{"# Proposal: Test Topic"}}` → `{"design", []string{"# Design: Test Topic"}}`

#### 3. cmd/rpi/scaffold.go
- Line 16: `proposalFlag string` → `designFlag string`
- Line 27: `"propose": "proposals"` → `"design": "designs"` in typeDirs
- Line 34: `"propose"` → `"design"` in validTypes
- Line 44: help text `propose → .rpi/proposals/` → `design → .rpi/designs/`
- Lines 74, 77-78: examples updated to use `design` and `.rpi/designs/`
- Line 93: `--proposal` flag → `--design` flag
- Line 123: `Proposal: proposalFlag` → `Design: designFlag`
- Lines 132: `"propose": "Proposal"` → `"design": "Design"` in local typeLabels
- Line 164: `case "...", "propose", ...` → `"design"` in validateRequiredFlags

#### 4. cmd/rpi/scaffold_test.go
- Line 191: `{"propose", ...}` → `{"design", ...}` and topic text updated

#### 5. cmd/rpi/serve.go
- Line 243: scanInput `Proposal` field → `Design`, jsonschema tag updated
- Line 245: scanInput jsonschema description: `"proposal"` → `"design"`
- Line 251: scaffoldInput type description: `"propose"` → `"design"`
- Line 255: scaffoldInput `Proposal` field → `Design`, jsonschema tag updated
- Line 331: `Proposal: input.Proposal` → `Design: input.Design`
- Line 354: `Proposal: input.Proposal` → `Design: input.Design`
- Line 360: `"propose": "Proposal"` → `"design": "Design"` in labels map

#### 6. cmd/rpi/scan.go
- Line 16: `scanProposal` → `scanDesign`
- Line 28: valid types help text `proposal` → `design`
- Line 39: example `.rpi/proposals/` → `.rpi/designs/`
- Line 57: `--type` description: `proposal` → `design`
- Line 58: `--proposal` flag → `--design` flag
- Line 68: `Proposal: scanProposal` → `Design: scanDesign`

#### 7. cmd/rpi/chain.go
- Line 21: link fields help text `proposal` → `design`
- Line 39: example `proposals` → `designs`, `"proposal"` → `"design"`

#### 8. cmd/rpi/archive.go
- Line 28: example `.rpi/proposals/` → `.rpi/designs/`
- Line 56: example `.rpi/proposals/` → `.rpi/designs/`

#### 9. cmd/rpi/extract.go
- Lines 30, 33, 36: examples `.rpi/proposals/` → `.rpi/designs/`

#### 10. cmd/rpi/main.go
- Line 20: `Propose` wording stays (it's the action verb), but check if "proposal" appears as artifact noun

#### 11. cmd/rpi/init_cmd.go
- Line 60: help text `proposals` → `designs`
- Line 159: `"proposals"` → `"designs"` in rpiSubdirs

#### 12. cmd/rpi/init_cmd_test.go
- Line 54: `"proposals"` → `"designs"` in expected subdirs
- Line 130 (and similar): verification of `proposals` dir → `designs`
- Line 358: `rpi-propose.md` stays (command name unchanged)

#### 13. internal/chain/resolve.go
- Line 15: `"proposal"` → `"design"` in linkFields
- Line 209-210: `case "proposals": return "proposal"` → `case "designs": return "design"` in inferType

#### 14. internal/chain/resolve_test.go
- Line 25: `proposalPath` → `designPath`, directory `proposals` → `designs`
- Line 32, 64, 66, 82: frontmatter `proposal:` → `design:` in test data
- Line 100: `proposalPath` → `designPath`, directory updated
- Line 127: directory `proposals` → `designs`
- Line 172: `"proposal: "` → `"design: "` in link construction
- Line 195: directory `proposals` → `designs`
- Line 254: `{".rpi/proposals/foo.md", "proposal"}` → `{".rpi/designs/foo.md", "design"}`

#### 15. internal/scanner/scan.go
- Line 25: `Proposal string` → `Design string` in Filters
- Line 112: `getStr(doc.Frontmatter, "proposal")` → `getStr(doc.Frontmatter, "design")`
- Line 111: `f.Proposal` → `f.Design`
- Line 244-245: `case "proposals": return "proposal"` → `case "designs": return "design"` in InferType

#### 16. internal/scanner/scan_test.go
- Line 26, 28: `"proposals/prop1.md"` → `"designs/prop1.md"`
- Line 31: `proposal: proposals/prop1.md` → `design: designs/prop1.md`
- Line 39: `"archive/proposals/old.md"` → `"archive/designs/old.md"`
- Line 81, 90: `Filters{Type: "proposal"}` → `Filters{Type: "design"}`
- Line 99: `Filters{Type: "proposal", Status: "draft"}` → `Filters{Type: "design", ...}`
- Line 115: `Filters{Proposal: "proposals/prop1.md"}` → `Filters{Design: "designs/prop1.md"}`
- Line 129: `Filters{References: "proposals/prop1.md"}` → `Filters{References: "designs/prop1.md"}`

### Success Criteria:

#### Automated Verification:
- [x] `go build ./...` succeeds
- [x] `go test ./...` — all tests pass
- [x] `grep -r "proposalFlag\|\"propose\"\|\"proposals\"\|\"proposal\"" --include="*.go" cmd/ internal/` returns no matches (excluding natural language in help text like "proposing")

#### Manual Verification:
- [x] `go run ./cmd/rpi scaffold design --topic "test" | head -5` shows `tags: [design]` and `# Design: test`
- [x] `go run ./cmd/rpi scaffold propose --topic "test"` errors with unknown type

### Commit:
- [x] Stage: `cmd/rpi/` `internal/`
- [x] Message: `refactor(artifacts): rename proposal type to design across Go code`

**Pause for manual confirmation before proceeding to next phase.**

---

## Phase 2: Template rename + restructure

### Overview
Rename `propose.tmpl` → `design.tmpl` with new section structure. Update `plan.tmpl` and `verify-report.tmpl` frontmatter field from `proposal:` to `design:`.

**Spec behaviors**: DA-1, DA-2, DA-3, DA-4, DA-5, DA-6, DA-14, DA-15

### Tasks:

#### 1. Rename and rewrite propose.tmpl → design.tmpl (3 locations)
**Files**:
- `internal/workflow/assets/templates/propose.tmpl` → `design.tmpl`
- `.claude/templates/propose.tmpl` → `design.tmpl`
- `.opencode/templates/propose.tmpl` → `design.tmpl`

**New content** (replaces entire file):
```
---
date: {{.Date}}
topic: "{{.Topic}}"
tags: [design{{if .Tags}}, {{.Tags}}{{end}}]
status: draft
{{- if .Research}}
related_research: {{.Research}}
{{- end}}
---

# Design: {{.Topic}}

## Summary
<!-- TODO: 1-3 sentence elevator pitch -->

## Context
<!-- TODO: Why this change — relevant background and constraints that shaped the design -->

## Constraints
<!-- TODO: What limits our options -->

## Components
<!-- TODO: What pieces exist or are being built, their responsibilities.
Integrate decisions inline — describe the component, mention alternatives
considered, explain the choice in context. -->

## File Structure
<!-- TODO: New/modified files — when applicable -->

## Risks
<!-- TODO: What could go wrong and how we'd handle it -->

## Out of Scope
<!-- TODO: What this design intentionally does not cover -->

## References
{{- if .Research}}
- Research: {{.Research}}
{{- end}}
```

#### 2. Update plan.tmpl frontmatter (3 locations)
**Files**:
- `internal/workflow/assets/templates/plan.tmpl`
- `.claude/templates/plan.tmpl`
- `.opencode/templates/plan.tmpl`

**Changes**: `proposal:` → `design:`, `{{.Proposal}}` → `{{.Design}}`, "Proposal" → "Design" in Source Documents and References sections

#### 3. Update verify-report.tmpl frontmatter (3 locations)
**Files**:
- `internal/workflow/assets/templates/verify-report.tmpl`
- `.claude/templates/verify-report.tmpl`
- `.opencode/templates/verify-report.tmpl`

**Changes**: `proposal:` → `design:`, `{{.Proposal}}` → `{{.Design}}`, "Proposal" → "Design" in References section

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` — all tests pass (template rendering tests use new template name)
- [x] `grep -r "propose\\.tmpl\|Proposal:" --include="*.tmpl"` returns no matches

#### Manual Verification:
- [x] `go run ./cmd/rpi scaffold design --topic "test" --write` creates file with correct sections: Summary, Context, Constraints, Components, File Structure, Risks, Out of Scope, References
- [x] Template does NOT contain Interfaces or Data Flow sections
- [x] `go run ./cmd/rpi scaffold plan --topic "test" --design ".rpi/designs/foo.md"` shows `design:` in frontmatter

### Commit:
- [x] Stage: `internal/workflow/assets/templates/`
- [x] Message: `refactor(templates): rename propose.tmpl to design.tmpl with restructured sections`

**Pause for manual confirmation before proceeding to next phase.**

---

## Phase 3: Command files

### Overview
Update `rpi-propose.md` to scaffold `type=design` and reference "design" artifacts. Update other command files (`rpi-plan.md`, `rpi-implement.md`, `rpi-verify.md`, `rpi-archive.md`) to replace "proposal" artifact references with "design".

**Spec behaviors**: DA-13 (command files reference design artifacts)

### Tasks:

#### 1. Update rpi-propose.md (3 locations)
**Files**:
- `internal/workflow/assets/commands/rpi-propose.md`
- `.claude/commands/rpi-propose.md`
- `.opencode/commands/rpi-propose.md`

**Changes** (throughout each file):
- "proposal" (as artifact noun) → "design" (e.g., "proposal document" → "design document")
- `type=propose` → `type=design` in scaffold calls
- `.rpi/proposals/` → `.rpi/designs/` in paths
- "Proposal:" heading → "Design:" heading
- Keep "propose" as the action verb (e.g., "propose solutions" stays)
- Keep command name `/rpi-propose` unchanged

#### 2. Update rpi-plan.md (3 locations)
**Files**:
- `internal/workflow/assets/commands/rpi-plan.md`
- `.claude/commands/rpi-plan.md`
- `.opencode/commands/rpi-plan.md`

**Changes**:
- "proposal" (artifact) → "design" throughout
- "proposal documents from `/rpi-propose`" → "design documents from `/rpi-propose`"
- "transition the proposal to complete" → "transition the design to complete"
- `.rpi/proposals/` → `.rpi/designs/` in paths

#### 3. Update rpi-implement.md (3 locations)
**Changes**: "proposals (`.rpi/proposals/`)" → "designs (`.rpi/designs/`)", "linked proposals" → "linked designs"

#### 4. Update rpi-verify.md (3 locations)
**Changes**: "proposal artifacts" → "design artifacts" throughout

#### 5. Update rpi-archive.md (3 locations) — if it has artifact references

### Success Criteria:

#### Automated Verification:
- [x] `grep -rn "\.rpi/proposals\|proposal document\|proposal artifact" --include="*.md" .claude/commands/ .opencode/commands/ internal/workflow/assets/commands/` returns no matches

#### Manual Verification:
- [x] Read `rpi-propose.md` — scaffolds `type=design`, suggests `→ /rpi-plan .rpi/designs/...`
- [x] Read `rpi-plan.md` — references "design documents", transitions "design" to complete

### Commit:
- [x] Stage: `internal/workflow/assets/commands/`
- [x] Message: `docs(commands): update artifact references from proposal to design`

**Pause for manual confirmation before proceeding to next phase.**

---

## Phase 4: Documentation templates + CLAUDE.md

### Overview
Update documentation templates (CLAUDE.md.template, AGENTS.md.template, PIPELINE.md.template) and the project root CLAUDE.md to reference `.rpi/designs/` and "design" artifacts.

### Tasks:

#### 1. Update CLAUDE.md.template (3 locations)
**Files**:
- `internal/workflow/assets/templates/CLAUDE.md.template`
- `.claude/templates/CLAUDE.md.template`
- `.opencode/templates/CLAUDE.md.template`

**Changes**:
- `.rpi/proposals/` → `.rpi/designs/` (or `.thoughts/designs/` for opencode)
- "Proposals: Record investigation findings..." → "Designs: Record architecture and design decisions..."
- "Proposals go in `.rpi/proposals/`" → "Designs go in `.rpi/designs/`"
- Pipeline reference: stays "Propose" (action verb) but artifact output is "design"
- Keep "proposing" as natural English verb where it's not referring to the artifact type (e.g., "before proposing commits" stays)

#### 2. Update AGENTS.md.template (3 locations)
**Files**:
- `internal/workflow/assets/templates/AGENTS.md.template`
- `.claude/templates/AGENTS.md.template`
- `.opencode/templates/AGENTS.md.template`

**Changes**: Same as CLAUDE.md.template

#### 3. Update PIPELINE.md.template (3 locations)
**Files**:
- `internal/workflow/assets/templates/PIPELINE.md.template`
- `.claude/templates/PIPELINE.md.template`
- `.opencode/templates/PIPELINE.md.template`

**Changes**:
- Diagram: keep "Propose" as stage name, output artifact becomes "design" + "spec"
- `.rpi/proposals/` → `.rpi/designs/`
- "proposal document" → "design document"
- Output section: file path updated to `.rpi/designs/...`
- Structure description: updated to reflect new template sections (Summary, Context, Constraints, Components, etc.)

#### 4. Update project root CLAUDE.md
**File**: `CLAUDE.md`

**Changes**:
- Line 24: `.rpi/proposals/` → `.rpi/designs/`
- Line 36: "Proposals: Record investigation findings..." → "Designs: Record architecture and design decisions..."
- Line 40: "Proposals go in `.rpi/proposals/`" → "Designs go in `.rpi/designs/`"
- Pipeline reference updated

### Success Criteria:

#### Automated Verification:
- [x] `grep -rn "proposals/" --include="*.md" --include="*.template" .claude/templates/ .opencode/templates/ internal/workflow/assets/templates/ CLAUDE.md | grep -v "proposing\|\.rpi/archive"` returns no matches

#### Manual Verification:
- [x] CLAUDE.md directory listing shows `.rpi/designs/`
- [x] PIPELINE.md.template stage name still says "Propose" but output artifact is "design"

### Commit:
- [x] Stage: `internal/workflow/assets/templates/*.template` `CLAUDE.md`
- [x] Message: `docs: update artifact references from proposal to design`

---

## References
- Proposal: .rpi/proposals/2026-03-19-rename-proposal-to-design.md
- Spec: .rpi/specs/design-artifact.md
