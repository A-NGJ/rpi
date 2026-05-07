---
domain: rpi-plan-split
feature: rpi-plan-split
last_updated: 2026-05-07T10:01:42+02:00
updated_by: .rpi/designs/2026-05-07-rpi-plan-split.md
---

# rpi-plan-split

## Purpose

When a design is too large for a single implementation plan, the planning step splits it into multiple sibling plans — each independently approvable, implementable, and reviewable — instead of producing one oversized plan. The split is automatic for fast-forward runs and proposed-then-confirmed in interactive runs; ordering between sibling plans is expressed as declared dependencies.

## Scenarios

### Simple design produces a single plan unchanged
Given a design whose structure does not signal complexity (one component cluster, one directory touched, one spec)
When the user invokes the planning step against that design
Then exactly one plan is produced for the design, with no split proposal shown, identical in shape to a plan produced before this feature existed.

### Complex design triggers a split proposal
Given a design whose structure crosses the complexity threshold (multiple distinct components, multiple directories touched, or multiple specs referenced)
When the user invokes the planning step against that design in interactive mode
Then before any plan file is written, the user is shown a labeled breakdown listing each proposed sibling plan, the design components it covers, the files it touches, and its declared dependencies on other proposed plans, together with the reasoning that triggered the split.

### User accepts the proposed split
Given a split proposal has been shown
When the user accepts the proposal as-is
Then one plan artifact is created per proposed sibling, each linked to the same upstream design, each declaring the dependencies shown in the proposal, and each surfaced to the user in the order it was proposed.

### User edits the split before accepting
Given a split proposal has been shown
When the user requests changes (regrouping components across plans, renaming a plan's scope, adding or removing a dependency between plans)
Then the breakdown is re-shown with the changes applied and validated as a directed acyclic graph; no plan file is written until the user accepts a re-shown breakdown.

### User declines the split and falls back to a single plan
Given a split proposal has been shown
When the user opts to plan as a single artifact instead
Then exactly one plan is produced for the design, identical in shape to today's single-plan output, with no leftover proposal state recorded.

### Fast-forward mode auto-splits and chains through implementation
Given a design crosses the complexity threshold and the user invokes planning with the fast-forward option
When the planning step runs without interactive gates
Then sibling plans are produced without a confirmation step, and implementation and verification proceed across the sibling plans in an order that respects each plan's declared dependencies.

### Resuming after a partial split is non-destructive
Given some sibling plans for a design already exist on disk from an earlier interrupted planning session
When the user invokes the planning step again against the same design
Then the existing sibling plans are detected and surfaced to the user, who chooses whether to write any remaining proposed plans or treat the split as complete; existing plan files are never silently overwritten or duplicated.

### Cyclic dependencies are rejected before any file is written
Given a proposed or user-edited split contains a cycle in its declared dependencies
When the planning step attempts to validate the breakdown
Then the user is shown which plans participate in the cycle, no plan files are written, and the user is given the chance to edit the breakdown before retrying.

## Constraints

- A design that does not cross the complexity threshold is planned as a single plan with no split proposal shown.
- In interactive mode, the user must accept a breakdown before any plan file is written; declining or aborting leaves no plan files on disk.
- Sibling plans share the same upstream design, and each can be implemented and merged independently of its siblings — completing one sibling never leaves the codebase in a broken state.
- Dependencies between sibling plans form a directed acyclic graph; cycles are rejected.
- Sibling plans for a design may only depend on other sibling plans of that same design.
- The upstream design transitions to its post-planning state only after every proposed sibling plan has been written.
- The fast-forward option and the grilling option remain mutually exclusive; under fast-forward the split applies without confirmation, under grilling the split proposal is grilled before user approval.
- Standalone planning mode (no upstream design) cannot produce a split.

## Out of Scope

- Automatic creation of tickets or issues in external trackers (Linear, GitHub, etc.).
- Dependencies between plans that belong to different designs.
- Splitting plans that already exist on disk (no resplit operation).
- Migrating historical single plans into split form.
- A learned or per-invocation language-model classifier for deciding when to split — the threshold is deterministic.
- Visual or graphical rendering of the dependency graph.
- Auto-merging completed plan branches.
- Any change to how individual phases inside a single plan are structured or named.
