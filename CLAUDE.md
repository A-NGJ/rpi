# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

This project follows Spec-Driven Development (SDD). Behavioral specs live in `.rpi/specs/` and serve as the source of truth for expected behavior. Always consult relevant specs before implementing or modifying features.

<!-- TODO: Add brief project description -->

## Git Workflow 

When committing changes, always ask the user which files/directories to include before proposing commits. Never assume all unstaged/staged changes should be committed.
Watch for uncommitted work that should be preserved. Suggest a commit (via `/rpi-commit`) when the user moves on to a different topic with completed changes still uncommitted, or when the working diff grows large enough that it risks becoming hard to review as a single commit.

## RPI Artifacts Directory

This project uses a `.rpi/` directory for persistent context:

```
.rpi/
├── research/      # Research notes (optional, from /rpi-research)
├── designs/       # Solution designs (created by /rpi-propose)
├── plans/         # Implementation plans (created by /rpi-plan)
├── specs/         # Living behavioral specs
├── reviews/       # Verification reports
├── diagnoses/     # Bug diagnosis post-mortems (created by /rpi-diagnose)
├── archive/       # Archived completed artifacts
```

### Development Pipeline

Workflow: Research → Propose → Plan → Implement → Verify

- **Research** (`/rpi-research`): Investigate the question (codebase or external). Optional.
- **Propose** (`/rpi-propose`): Analyze trade-offs, write design + spec (behavioral contract). Approval gate. A read-only pre-lock audit checks the drafted Components cohere (coverage, cross-Component mismatch, decision-drift) before the gate. Carries upstream decisions forward (via `rpi chain --sections "Decisions"`) into an `## Inherited Decisions` block, each entry attributed to its source artifact.
- **Plan** (`/rpi-plan`): Create phased implementation plan from approved spec. A read-only pre-lock audit checks the drafted phases cohere (coverage, forward-references, decision-drift) before the buy-in gate. Inherits upstream decisions (design and, transitively, research) into an `## Inherited Decisions` block with per-source attribution.
- **Blueprint** (`/rpi-blueprint`): Fused shortcut for low-stakes solo work — research note or short problem statement → phased plan in one pass, omitting the standalone design deliverable but still emitting a minimal spec (the dropped design reasoning is captured in a `## Design Notes` plan block). Distinct from `--ff` (full pipeline fast, still produces a design) and from `/rpi-plan` (scoped change, no design reasoning). Refuses and redirects to `/rpi-propose` on genuine tradeoffs or high blast radius. Optional.
- **Implement** (`/rpi-implement`): Execute plan phase-by-phase with verification.
- **Revise** (`/rpi-revise`): Amend an existing plan when a new constraint or a review finding lands after it was drafted or partially implemented — edits only the affected phases, preserves the checkbox state of completed work, re-audits only what changed, then hands back to implement. Distinct from `/rpi-plan` (amend an existing plan vs. create a fresh one) and closes the verify → revise → implement loop. Optional.
- **Verify** (`/rpi-verify`): Validate spec conformance. A read-only grounding pass re-anchors each finding against repo state and demotes blockers it can't confirm.
- **Diagnose** (`/rpi-diagnose`): Iterative root-cause analysis and fix for complex bugs. Optional.
- **Explain** (`/rpi-explain`): Diff-scoped walkthrough of an implemented solution. Optional.

Each command suggests the next step. Start with `/rpi-propose` for features, `/rpi-plan` for bug fixes, `/rpi-blueprint` for low-stakes solo work that wants a plan without a separate design, `/rpi-diagnose` for complex bugs, `/rpi-research` when exploring.

## Codebase Navigation

When exploring unfamiliar code, check what navigation tools are available before falling back to text search. Structural overviews and definition lookups are more efficient than scanning files when you need to understand how a codebase is organized or where something is defined.

## Development Conventions

Before implementing any changes, always: 1) Read the current version of each file you plan to modify, 2) Run the existing test suite to establish a baseline, 3) Implement changes incrementally — one logical unit at a time, 4) Run tests after each unit. If tests fail, fix before proceeding. Do not batch all changes and test at the end.
<!-- TODO: Add project-specific conventions -->

When implementing a plan from `.rpi/plans/`, present intended changes for each phase before writing code. If a phase's success criteria are fully covered by automated checks (tests, linting, etc.), run them and proceed automatically when they pass. Only pause for manual verification when the plan includes manual verification items not covered by automated tests. Update checkboxes in the plan file as items complete, and resume from the first unchecked item if checkboxes already exist.

