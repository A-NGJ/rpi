---
date: 2026-04-04T22:16:33+02:00
status: active
tags:
    - design
topic: gherkin-inspired spec format and spec-aware verification
---

# Design: Gherkin-Inspired Spec Format and Spec-Aware Verification

## Summary

Replace the current XX-N requirement list spec format with a Gherkin-inspired scenario format that targets ~5-8 behavioral scenarios per spec. Extend `rpi verify` with a `spec` subcommand that parses scenarios and produces structured pass/fail output. Update skill prompts to generate and verify against the new format.

## Context

Current specs have 9-26 requirements each, using prefixed IDs (XX-N) with a mix of behavioral, structural, constraint, and implementation-detail requirements. As specs grow, they drift from user-observable behavior toward implementation details (e.g., "The Index struct includes an `Imports []Import` field"). This makes specs hard to maintain, review, and verify.

The current `rpi verify` CLI only checks plan checkboxes and TODO markers — it never reads specs. Actual spec verification is done entirely by Claude reading code manually during `/rpi-verify`, with no structured tooling support.

Industry research identified **Gherkin/BDD** and **Specification by Example** as the best-fit formats for a dual human+AI audience where automated verification matters. Both emphasize observable behavior through concrete examples and naturally prevent over-specification.

### What's tightly coupled to XX-N (only 3 files)

1. `internal/workflow/assets/templates/spec.tmpl` — template with XX-N examples
2. `.claude/skills/rpi-propose/SKILL.md` line 27 — invariant requiring XX-N
3. `.claude/skills/rpi-verify/SKILL.md` line 20 — invariant checking XX-N

Everything else (scaffold, scan, frontmatter, template rendering, scanner) is format-agnostic.

## Constraints

- Must work for both human reviewers and Claude as implementer/verifier
- Must not require external tooling (no Cucumber, no step definitions)
- Must not require bulk migration of existing specs — coexistence during transition
- Verify CLI output must be structured (JSON) for programmatic consumption

## Components

### 1. New Spec Template

Replace the current template with a Gherkin-inspired markdown format. The key design decision is **scenarios as the primary unit** instead of numbered requirements.

**Alternative considered**: Pure Gherkin `.feature` files — rejected because they require Cucumber tooling and step definition glue code. Markdown keeps specs readable, editable, and renderable without extra tools.

**Alternative considered**: Keep XX-N but add a "max 8 requirements" guideline — rejected because the format itself encourages granularity. Numbered lists invite completionism; scenarios invite behavioral thinking.

New template structure:

```markdown
---
domain: {{.Topic}}
feature: <!-- TODO: short feature name, e.g. rpi-status -->
last_updated: {{.Date}}
updated_by: <!-- TODO: design or plan that created/updated this -->
---

# {{.Topic}}

## Purpose
<!-- TODO: What this feature does and why — 1-3 sentences -->

## Scenarios

### <!-- TODO: Scenario title (verb phrase describing the behavior) -->
Given <!-- precondition -->
When <!-- action -->
Then <!-- observable outcome -->

## Constraints
- <!-- TODO: Boundaries and invariants -->

## Out of Scope
- <!-- TODO: What this spec intentionally does not cover -->
```

Key changes from current format:
- `id` field (XX prefix) replaced with `feature` field (descriptive name)
- `## Behavior` with XX-N lists replaced with `## Scenarios` using Given/When/Then
- `## Test Cases` section removed — scenarios *are* the test cases
- `### Must` / `### Must Not` collapsed into flat `## Constraints`
- Target: 5-8 scenarios per spec

### 2. Verify Spec Subcommand

New CLI subcommand: `rpi verify spec <spec-path>`

Parses the `## Scenarios` section, extracts individual scenario blocks (title + Given/When/Then steps), and returns structured JSON.

```json
{
  "spec": ".rpi/specs/rpi-status.md",
  "feature": "rpi-status",
  "scenarios": [
    {
      "title": "Display artifact summary",
      "given": "the .rpi/ directory contains 3 active plans and 2 draft designs",
      "when": "the user runs `rpi status`",
      "then": "output shows artifacts grouped by type with status counts"
    }
  ],
  "total": 6
}
```

This gives the Claude skill structured data to verify against, rather than relying on the LLM to parse markdown itself. The CLI parses; Claude verifies each scenario against code.

**Alternative considered**: Have the CLI also run verification (check code against scenarios) — rejected because behavioral verification requires semantic understanding that only the LLM can provide. The CLI's job is structured parsing; Claude's job is semantic verification.

Implementation leverages existing `internal/frontmatter/sections.go` for section extraction and adds scenario-block parsing on top.

### 3. Skill Prompt Updates

**rpi-propose** (`.claude/skills/rpi-propose/SKILL.md`):
- Change invariant from "Create a behavioral spec with prefixed IDs (XX-N)" to "Create a behavioral spec with 5-8 Given/When/Then scenarios describing observable behavior"
- Add guidance: scenarios must describe user-observable behavior, not internal structure

**rpi-verify** (`.claude/skills/rpi-verify/SKILL.md`):
- Change from "verify spec behaviors (XX-N)" to "use `rpi verify spec` to extract scenarios, then verify each scenario against actual code"
- Verification report format stays the same (completeness/correctness/coherence with severity classification)

### 4. MCP Tool

Register `rpi_verify_spec` as a new MCP tool alongside existing `rpi_verify_completeness` and `rpi_verify_markers`. Input: `spec_path` (string). Output: parsed scenario JSON.

## File Structure

| File | Action | Description |
|------|--------|-------------|
| `internal/workflow/assets/templates/spec.tmpl` | Modify | New Gherkin-inspired template |
| `cmd/rpi/verify.go` | Modify | Add `spec` subcommand with scenario parsing |
| `cmd/rpi/serve.go` | Modify | Register `rpi_verify_spec` MCP tool |
| `.claude/skills/rpi-propose/SKILL.md` | Modify | Update invariant to scenario format |
| `.claude/skills/rpi-verify/SKILL.md` | Modify | Update to use `rpi verify spec` |

## Risks

**Scenario quality depends on the writer.** Bad scenarios (too vague or too implementation-coupled) produce the same problems as bad XX-N requirements. Mitigation: skill prompt includes explicit guidance on what makes a good scenario (observable behavior, concrete examples, no internal structure).

**Existing specs use XX-N format.** During transition, verify needs to handle both formats. Mitigation: the new `rpi verify spec` subcommand only handles the new format. Old specs continue to be verified via the existing manual skill flow until migrated.

### 5. Bulk Migration of Existing Specs

After all tooling and skill prompts are updated, migrate existing XX-N specs in `.rpi/specs/` to the new scenario format. Each spec is rewritten with 5-8 behavioral scenarios that preserve the same domain coverage and constraints. This is the final step — it removes the need for dual-format support.

## Out of Scope

- Executable test generation from scenarios (future possibility)
- Scenario tagging or categorization beyond the feature field
- Changes to non-spec artifacts (plans, designs, research)

## References

- Gherkin Reference: https://cucumber.io/docs/gherkin/reference/
- Specification by Example (Gojko Adzic)
- Current spec template: `internal/workflow/assets/templates/spec.tmpl`
