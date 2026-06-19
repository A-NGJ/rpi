---
domain: skills
feature: rpi-revise
last_updated: 2026-06-17T10:00:00+02:00
updated_by: .rpi/designs/2026-06-17-rpi-revise-skill.md
---

# rpi-revise

## Purpose

Define the behavioral contract for `rpi-revise`: amending an existing implementation plan when a new constraint or a review finding lands after the plan is drafted or partially implemented. Covers when the skill fires, how already-completed work is protected, how revision scopes its re-audit to the changed parts of the plan, how it behaves on finished plans, and how it hands back to implementation. The defining promise is that revising a plan never silently undoes work that was already done, and never re-checks parts of the plan the change did not touch.

## Scenarios

### Revising a partially-implemented plan preserves completed work
Given a developer has a plan whose early phases are already done and checked off, and a later phase is still pending
When they ask to revise the plan to account for a new requirement that affects only the pending phase
Then the completed phases stay marked done, only the pending phase is amended, and the developer is shown that no finished work was reverted

### A revision is scoped to the parts of the plan it touches
Given a plan with several phases
When the developer revises it for a change that affects one phase
Then only that phase is re-examined and rewritten, the other phases are left exactly as they were, and the developer is told which phase changed before anything is written

### The skill refuses to silently undo completed work
Given a plan with completed, checked-off work
When a proposed revision would reset some of that completed work back to not-done
Then the skill stops, shows the developer exactly which completed items the change would reopen, and asks for explicit confirmation rather than writing the change silently

### A constraint that arrives mid-implementation amends the active plan in place
Given a developer is partway through implementing an active plan when a new constraint emerges
When they describe the constraint and ask to revise
Then the plan is updated in place, its in-progress status is preserved, and the skill suggests resuming implementation from the first item that is not yet done

### A review finding flows back into the plan and then to implementation
Given a verification of a finished or in-flight implementation surfaces a gap
When the developer asks to revise the plan to close that gap
Then the affected phase is amended to cover the gap, unaffected phases keep their state, and the skill suggests re-running implementation to pick up the new work — closing the verify-then-revise-then-implement loop without starting a fresh design

### Revising a finished plan requires an explicit choice
Given a plan that is already marked complete
When the developer asks to revise it
Then the skill does not silently reopen the finished plan; instead it explains that the plan is complete and offers two explicit choices — reopen it deliberately, or supersede it with a fresh successor plan that carries the unchanged work forward

### Adding a new phase keeps earlier completed phases intact
Given a plan with its first two phases checked off
When the developer revises it to insert an additional phase of work after them
Then the first two phases remain checked, the new phase is added as not-yet-done, and resuming implementation begins at the newly added work rather than redoing the earlier phases

### Fast-forward revision still protects completed work
Given a developer revises a partially-done plan in fast-forward mode to skip the approval step
When the revision is applied automatically and implementation is resumed automatically
Then the approval pause is skipped, but the protection against silently undoing completed work is still enforced — a revision that would reopen finished work stops and reports instead of proceeding

## Constraints

- A revision never silently changes a done item back to not-done. Any reversal of completed work requires explicit developer confirmation, in every mode including fast-forward.
- A revision re-examines and rewrites only the phases the change affects. Phases the change does not touch are left unchanged, including their done / not-done state.
- The skill fires to amend an *existing* plan for a new constraint or a review finding. A fresh, self-contained scoped change is the planning skill's job, and the description steers those away.
- After a successful revision the skill suggests resuming implementation from the first not-yet-done item.
- An in-progress plan stays in-progress across a revision; a draft plan stays a draft. A completed plan is never reopened without an explicit developer choice between reopening and superseding.
- Fast-forward and grill behavior follow the shared cross-skill mode contract; the two modes are mutually exclusive. Grill, when available, stress-tests the proposed amendment before it is written.
- The skill amends plans only. A change that invalidates the underlying design routes back through the design stage rather than being absorbed as a plan revision.

## Out of Scope

- Creating a brand-new plan from scratch — that remains the planning skill's responsibility.
- Editing or re-deriving the upstream design or spec. Revision operates on the plan layer only.
- Splitting a plan into sibling plans during a revision. If a revision grows the plan past the splitting threshold, the skill surfaces that and defers to the planning skill's split flow.
- Running the independent verification pass as part of a revision. Revision hands back to implementation; verification is run separately.
- A separate revision-history artifact. Revisions are edits to the existing plan, and the version-control history is the record.
- Any new command-line subcommand. The skill reuses existing capabilities for progress inspection, status transitions, and context recovery.
