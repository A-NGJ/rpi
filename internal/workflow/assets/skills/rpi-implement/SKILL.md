---
name: rpi-implement
description: Implement technical plans from .rpi/plans with verification
---

# Implement Plan

## Goal

Execute an active plan from `.rpi/plans/` phase by phase. Plans come in two forms: pipeline plans (reference designs and research) or standalone plans (self-contained).

After all phases are complete and verified, announce completion and update the plan status.

## Invariants

- Validate the plan's status before starting — draft/active: proceed; complete: warn about duplication
- Resolve the plan's artifact chain and read all upstream context (designs, research, specs)
- Read all files mentioned in the plan fully — never use limit/offset
- Transition the plan to active status when starting
- Check current progress (completed vs remaining items) — resume from first unchecked item
- **Pre-review**: before writing code for a phase, present all intended changes for approval — flag any deviations from plan
- **Red/green TDD**: for new code, write tests first (confirm they fail), then implement until they pass
- Run success criteria checks after each phase — fix issues before proceeding
- Update checkboxes in the plan file as items complete
- **Before committing**: scan staged files for sensitive content — warn and exclude if flagged
- Present commit plan (files + message) and ask for approval before executing
- **After hook failure**: read error, fix the issue, re-stage, create a new commit (never amend)
- If a phase's success criteria are fully covered by automated checks (tests, linting, etc.), run them and proceed automatically when they pass — only pause for manual verification when the plan includes manual verification items not covered by automated tests
- **On mismatch**: stop, present what the plan says vs what you found, ask how to proceed
- **On completion**: verify spec conformance for all linked specs — check each behavior (XX-N) against actual code and tests; plan → complete

## Principles

- Follow the plan's intent while adapting to what you find — your judgment matters
- Implement each phase fully before moving to the next
- Trust completed checkmarks when resuming — verify only if something seems off
