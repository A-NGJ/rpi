---
domain: spec-format
feature: spec-format
last_updated: 2026-04-04T22:16:33+02:00
updated_by: .rpi/designs/2026-04-04-gherkin-inspired-spec-format-and-spec-aware-verification.md
---

# Gherkin-Inspired Spec Format and Spec-Aware Verification

## Purpose

Specs use a Gherkin-inspired scenario format with 5-8 Given/When/Then scenarios per spec, replacing the XX-N numbered requirement lists. The `rpi verify spec` CLI subcommand parses scenarios into structured JSON so that both tooling and Claude can verify implementations against behavioral contracts.

## Scenarios

### New specs use scenario format
Given a user scaffolds a new spec with `rpi scaffold spec --topic "my feature"`
When the template is rendered
Then the output contains `## Scenarios` with Given/When/Then placeholders and a `feature` frontmatter field instead of `id`

### Scenarios describe observable behavior
Given a spec is created via `/rpi-propose`
When Claude writes the scenarios
Then each scenario describes user-observable behavior with concrete Given/When/Then steps and the spec contains 5-8 scenarios

### Verify parses scenarios from a spec
Given a spec file with a `## Scenarios` section containing named scenario blocks
When the user runs `rpi verify spec <spec-path>`
Then the CLI outputs JSON with the feature name, an array of parsed scenarios (title, given, when, then), and a total count

### Verify spec is available as MCP tool
Given the MCP server is running
When a client calls `rpi_verify_spec` with a spec path
Then it returns the same structured JSON as the CLI subcommand

### Skill prompt generates scenario-based specs
Given a user invokes `/rpi-propose` for a new feature
When Claude creates the behavioral spec
Then the spec uses the scenario format with `## Scenarios` section, not XX-N numbered requirements

### Skill prompt verifies against scenarios
Given a user invokes `/rpi-verify` on an implementation with a scenario-based spec
When Claude performs verification
Then each scenario is checked individually against actual code with a pass/fail per scenario and file:line references

### Old specs remain valid during transition
Given existing specs use the XX-N requirement format
When the user runs `/rpi-verify` against them
Then Claude verifies using the existing manual approach without errors

### Existing specs are migrated to scenario format
Given all new tooling and skill prompts use the scenario format
When the migration is performed on existing XX-N specs in `.rpi/specs/`
Then each spec is rewritten with 5-8 behavioral scenarios replacing its XX-N requirements, preserving the same domain coverage and constraints

## Constraints
- Spec template must be plain markdown — no Cucumber tooling or step definition files required
- `rpi verify spec` output must be valid JSON for programmatic consumption
- Scenario parsing must handle multi-line Given/When/Then steps (lines continuing without a new keyword)
- Existing XX-N specs must not break — both formats coexist during transition
- No changes to non-spec artifact formats (plans, designs, research, diagnoses)

## Out of Scope
- Automated test generation from scenarios
- Scenario tagging, prioritization, or categorization
- Changes to the plan, design, or research templates
